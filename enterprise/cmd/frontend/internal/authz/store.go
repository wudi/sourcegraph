package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Store is the unified interface for managing permissions explicitly in the database.
// It is concurrent-safe and maintains the data consistency over 'user_permissions',
// 'repo_permissions' and 'user_pending_permissions' tables.
type Store struct {
	db    dbutil.DB
	clock func() time.Time
}

// NewStore returns a new Store with given parameters.
func NewStore(db dbutil.DB, clock func() time.Time) *Store {
	return &Store{
		db:    db,
		clock: clock,
	}
}

// LoadUserPermissions loads stored user permissions into p. An error is returned
// when there are no valid permissions available.
func (s *Store) LoadUserPermissions(ctx context.Context, p *UserPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.load(ctx, loadUserPermissionsQuery(p))
	return err
}

func loadUserPermissionsQuery(p *UserPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadUserPermissionsQuery
SELECT object_ids, updated_at
FROM user_permissions
WHERE user_id = %s
AND permission = %s
AND object_type = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		p.Provider,
	)
}

// LoadRepoPermissions loads stored repository permissions into p. An error is
// returned when there are no valid permissions available.
func (s *Store) LoadRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.UserIDs, p.UpdatedAt, err = s.load(ctx, loadRepoPermissionsQuery(p))
	return err
}

func loadRepoPermissionsQuery(p *RepoPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:loadRepoPermissionsQuery
SELECT user_ids, updated_at
FROM repo_permissions
WHERE repo_id = %s
AND permission = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		p.Provider,
	)
}

// LoadPendingPermissions loads stored pending user permissions into p. An
// error is returned when there are no pending permissions available.
func (s *Store) LoadPendingPermissions(ctx context.Context, p *PendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadPendingPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.load(ctx, loadPendingPermissionsQuery(p))
	return err
}

func loadPendingPermissionsQuery(p *PendingPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:loadPendingPermissionsQuery
SELECT object_ids, updated_at
FROM user_pending_permissions
WHERE bind_id = %s
AND permission = %s
AND object_type = %s
`

	return sqlf.Sprintf(
		format,
		p.BindID,
		p.Perm.String(),
		p.Type,
	)
}

// SetRepoPermissions performs a full update for p, new IDs found in p will be upserted
// and IDs no longer in p will be removed. This method updates both the user and
// repository permissions tables.
func (s *Store) SetRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Retrieve currently stored IDs of this repository.
	oldIDs, _, err := s.load(ctx, loadRepoPermissionsQuery(p))
	if err != nil {
		return err
	}

	// Fisrt get the intersection (And), then use the intersection to compute diffs (AndNot)
	// with both the old and new sets to get IDs to remove and to add.
	isec := p.UserIDs.Clone()
	isec.And(oldIDs)
	toRemove := isec.Clone()
	toRemove.AndNot(oldIDs)
	toAdd := isec.Clone()
	toAdd.AndNot(p.UserIDs)

	// Open a transaction for update consistency.
	var tx *sqlTx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}
	defer tx.commitOrRollback(&err)

	// Make another Store with this underlying transaction.
	txs := &Store{db: tx, clock: s.clock}

	var q *sqlf.Query

	err = txs.iterateAndUpsertUserPermissions(ctx, toAdd, p.Perm, PermRepos, p.Provider,
		func(set *roaring.Bitmap) bool {
			return set.CheckedAdd(uint32(p.RepoID))
		})
	if err != nil {
		return err
	}

	err = txs.iterateAndUpsertUserPermissions(ctx, toRemove, p.Perm, PermRepos, p.Provider,
		func(set *roaring.Bitmap) bool {
			return set.CheckedRemove(uint32(p.RepoID))
		})
	if err != nil {
		return err
	}

	isec.Or(toAdd)
	p.UserIDs = isec

	p.UpdatedAt = txs.clock()
	if q, err = upsertRepoPermissionsQuery(p); err != nil {
		return err
	} else if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	return nil
}

func upsertRepoPermissionsQuery(p *RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertRepoPermissionsQuery
INSERT INTO repo_permissions
  (repo_id, permission, user_ids, provider, updated_at)
VALUES
  (%s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  repo_permissions_perm_provider_unique
DO UPDATE SET
  user_ids = excluded.user_ids,
  updated_at = excluded.updated_at
`

	p.UserIDs.RunOptimize()
	ids, err := p.UserIDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		ids,
		p.Provider,
		p.UpdatedAt.UTC(),
	), nil
}

// iterateAndUpsertUserPermissions uses the iterator to check if the user permissions loaded
// for the user ID needs an upsert. The iterator is meant to return user IDs.
//
// The update is the check and modify function that determines whether an update to the user
// permissions is needed. It should return true if the set provided is being modified during
// the check.
func (s *Store) iterateAndUpsertUserPermissions(
	ctx context.Context,
	ids *roaring.Bitmap,
	perm authz.Perms,
	typ PermType,
	provider ProviderType,
	update func(set *roaring.Bitmap) bool,
) (err error) {
	_, save := s.observe(ctx, "iterateAndUpsertUserPermissions", "")
	defer func() {
		save(&err,
			otlog.String("perm", string(perm)),
			otlog.String("type", string(typ)),
			otlog.String("provider", string(provider)),
		)
	}()

	// Batch query all user permissions object IDs in one go.
	q := loadUserPermissionsBatchQuery(ids.ToArray(), perm, typ, provider)
	loadedIDs, err := s.batchLoad(ctx, q)
	if err != nil {
		return err
	}

	updatedAt := s.clock()
	iter := ids.Iterator()
	for iter.HasNext() {
		userID := int32(iter.Next())
		oldIDs := loadedIDs[userID]
		if oldIDs == nil {
			oldIDs = roaring.NewBitmap()
		}

		up := &UserPermissions{
			UserID:   userID,
			Perm:     perm,
			Type:     PermRepos,
			IDs:      oldIDs,
			Provider: provider,
		}
		if !update(up.IDs) {
			continue
		}

		up.UpdatedAt = updatedAt
		if q, err = upsertUserPermissionsQuery(up); err != nil {
			return err
		} else if err = s.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func loadUserPermissionsBatchQuery(
	userIDs []uint32,
	perm authz.Perms,
	typ PermType,
	provider ProviderType,
) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadUserPermissionsBatchQuery
SELECT user_id, object_ids
FROM user_permissions
WHERE user_id IN (%s)
AND permission = %s
AND object_type = %s
AND provider = %s
`

	items := make([]*sqlf.Query, len(userIDs))
	for i := range userIDs {
		items[i] = sqlf.Sprintf("%d", userIDs[i])
	}
	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
		perm,
		typ,
		provider,
	)
}

func upsertUserPermissionsQuery(p *UserPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertUserPermissionsQuery
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids, provider, updated_at)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_provider_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	p.IDs.RunOptimize()
	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		ids,
		p.Provider,
		p.UpdatedAt.UTC(),
	), nil
}

// SetPendingPermissions performs a full update for ps to the pending permissions table.
func (s *Store) SetPendingPermissions(ctx context.Context, ps ...*PendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetPendingPermissions", "")
	defer func() { save(&err) }()

	var q *sqlf.Query
	updatedAt := s.clock()
	for _, p := range ps {
		p.UpdatedAt = updatedAt
		if q, err = upsertPendingPermissionsQuery(p); err != nil {
			return err
		} else if err = s.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func upsertPendingPermissionsQuery(p *PendingPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertPendingPermissionsQuery
INSERT INTO user_pending_permissions
  (bind_id, permission, object_type, object_ids, updated_at)
VALUES
  (%s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_pending_permissions_perm_object_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	p.IDs.RunOptimize()
	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		format,
		p.BindID,
		p.Perm.String(),
		p.Type,
		ids,
		p.UpdatedAt.UTC(),
	), nil
}

// GrantPendingPermissions grants the user has given ID with pending permissions found in p.
func (s *Store) GrantPendingPermissions(ctx context.Context, userID int32, p *PendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "GrantPendingPermissions", "")
	defer func() {
		save(&err,
			append(p.TracingFields(), otlog.Object("userID", userID))...,
		)
	}()

	p.IDs, p.UpdatedAt, err = s.load(ctx, loadPendingPermissionsQuery(p))
	if err != nil {
		return err
	}

	// Open a transaction for update consistency.
	var tx *sqlTx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}
	defer tx.commitOrRollback(&err)

	// Make another Store with this underlying transaction.
	txs := &Store{db: tx, clock: s.clock}

	up := &UserPermissions{
		UserID:    userID,
		Perm:      p.Perm,
		Type:      p.Type,
		IDs:       p.IDs,
		Provider:  ProviderSourcegraph,
		UpdatedAt: txs.clock(),
	}
	var q *sqlf.Query
	if q, err = upsertUserPermissionsQuery(up); err != nil {
		return err
	} else if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	// NOTE: We currently only have "repos" type, so avoid unnecessary type checking for now.

	// Batch query all repository permissions object IDs in one go.
	q = loadRepoPermissionsBatchQuery(p.IDs.ToArray(), p.Perm, ProviderSourcegraph)
	loadedIDs, err := s.batchLoad(ctx, q)
	if err != nil {
		return err
	}

	updatedAt := txs.clock()
	iter := p.IDs.Iterator()
	for iter.HasNext() {
		repoID := int32(iter.Next())
		oldIDs := loadedIDs[repoID]
		if oldIDs == nil {
			oldIDs = roaring.NewBitmap()
		}

		rp := &RepoPermissions{
			RepoID:   repoID,
			Perm:     p.Perm,
			UserIDs:  oldIDs,
			Provider: ProviderSourcegraph,
		}
		if !rp.UserIDs.CheckedAdd(uint32(userID)) {
			continue
		}

		rp.UpdatedAt = updatedAt
		if q, err = upsertRepoPermissionsQuery(rp); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func loadRepoPermissionsBatchQuery(
	repoIDs []uint32,
	perm authz.Perms,
	provider ProviderType,
) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadRepoPermissionsBatchQuery
SELECT repo_id, user_ids
FROM repo_permissions
WHERE repo_id IN (%s)
AND permission = %s
AND provider = %s
`

	items := make([]*sqlf.Query, len(repoIDs))
	for i := range repoIDs {
		items[i] = sqlf.Sprintf("%d", repoIDs[i])
	}
	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
		perm,
		provider,
	)
}

func (s *Store) upsert(ctx context.Context, q *sqlf.Query) (err error) {
	ctx, save := s.observe(ctx, "upsert", "")
	defer func() { save(&err, otlog.Object("q", q)) }()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

// load runs the query and returns unmarshalled IDs and last updated time.
func (s *Store) load(ctx context.Context, q *sqlf.Query) (*roaring.Bitmap, time.Time, error) {
	var err error
	ctx, save := s.observe(ctx, "load", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, time.Time{}, err
	}

	if !rows.Next() {
		return nil, time.Time{}, rows.Err()
	}

	var ids []byte
	var updatedAt time.Time
	if err = rows.Scan(&ids, &updatedAt); err != nil {
		return nil, time.Time{}, err
	}

	if err = rows.Close(); err != nil {
		return nil, time.Time{}, err
	}

	bm := roaring.NewBitmap()
	if len(ids) == 0 {
		return bm, time.Time{}, nil
	} else if err = bm.UnmarshalBinary(ids); err != nil {
		return nil, time.Time{}, err
	}

	return bm, updatedAt, nil
}

// batchLoad runs the query and returns unmarshalled IDs with their corresponding object ID value.
func (s *Store) batchLoad(ctx context.Context, q *sqlf.Query) (map[int32]*roaring.Bitmap, error) {
	var err error
	ctx, save := s.observe(ctx, "batchLoad", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}

	loaded := make(map[int32]*roaring.Bitmap)
	for rows.Next() {
		var objID int32
		var ids []byte
		if err = rows.Scan(&objID, &ids); err != nil {
			return nil, err
		}

		if len(ids) == 0 {
			continue
		}

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return nil, err
		}
		loaded[objID] = bm
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	return loaded, nil
}

func (s *Store) tx(ctx context.Context) (*sqlTx, error) {
	switch t := s.db.(type) {
	case *sql.Tx:
		return &sqlTx{t}, nil
	case *sql.DB:
		tx, err := t.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &sqlTx{tx}, nil
	default:
		panic(fmt.Sprintf("can't open transaction with unknown implementation of dbutil.DB: %T", t))
	}
}

func (s *Store) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "authz.Store."+family, title)

	return ctx, func(err *error, fs ...otlog.Field) {
		now := s.clock()
		took := now.Sub(began)

		fs = append(fs, otlog.String("Duration", took.String()))

		tr.LogFields(fs...)

		success := err == nil || *err == nil
		if !success {
			tr.SetError(*err)
		}

		tr.Finish()
	}
}
