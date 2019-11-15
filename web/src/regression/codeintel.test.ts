/**
 * @jest-environment node
 */

import { Driver } from '../../../shared/src/e2e/driver'
import { ensureLoggedInOrCreateTestUser } from './util/helpers'
import { getConfig } from '../../../shared/src/e2e/config'
import { getTestTools } from './util/init'
import { getUser, setUserSiteAdmin, ensureTestExternalService } from './util/api'
import { GraphQLClient } from './util/GraphQLClient'
import * as GQL from '../../../shared/src/graphql/schema'
import { TestResourceManager } from './util/TestResourceManager'
import { CodeNavTestCase, testCodeIntel, clearDumps, enableLSIF, uploadAndEnsureDump } from './util/codeintel'

describe('Code intelligence regression test suite', () => {
    const testUsername = 'test-sg-codeintel'
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'gitHubToken',
        'sourcegraphBaseUrl',
        'noCleanup',
        'testUserPassword',
        'keepBrowser',
        'logBrowserConsole',
        'slowMo',
        'headless',
        'keepBrowser',
        'logStatusMessages'
    )
    const testExternalServiceInfo = {
        kind: GQL.ExternalServiceKind.GITHUB,
        uniqueDisplayName: '[TEST] GitHub (codeintel.test.ts)',
    }
    const testRepoSlugs = ['go-nacelle/process']

    let driver: Driver
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
        resourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        resourceManager.add(
            'External service',
            testExternalServiceInfo.uniqueDisplayName,
            await ensureTestExternalService(
                gqlClient,
                {
                    ...testExternalServiceInfo,
                    config: {
                        url: 'https://github.com',
                        token: config.gitHubToken,
                        repos: testRepoSlugs,
                        repositoryQuery: ['none'],
                    },
                    waitForRepos: testRepoSlugs.map(slug => 'github.com/' + slug),
                },
                config
            )
        )
        const user = await getUser(gqlClient, testUsername)
        if (!user) {
            throw new Error(`test user ${testUsername} does not exist`)
        }
        await setUserSiteAdmin(gqlClient, user.id, true)
    }, 60 * 1000)
    afterAll(async () => {
        if (!config.noCleanup) {
            await resourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    }, 60 * 1000)

    test(
        'Dumps',
        async () => {
            // TODO - ensure repo is cloned
            const repository = 'github.com/go-nacelle/process'
            const commit = '1e2e6b62d8a4139e0101433bb1bd08db8bb3fa8f'
            const root = '/'
            const filename = 'nacelle-process.lsif'

            const testCases: CodeNavTestCase[] = [
                {
                    repoRev: `${repository}@${commit}`,
                    files: [
                        {
                            // TODO - ensure LSIF and not basic somehow

                            path: '/process.go',
                            locations: [
                                {
                                    line: 10,
                                    token: 'Initializer',
                                    expectedHoverContains:
                                        'Initializer is an interface that is called once on app startup.',
                                    expectedDefinition: `/${repository}@${commit}/-/blob/initializer.go#L8:2`,
                                    expectedReferences: [
                                        '/initializer.go#L8:2',
                                        '/parallel_initializer.go#L41:14',
                                        '/process.go#L10:3',
                                        '/process_container.go#L10:23',
                                        '/process_container.go#L55:14',
                                        '/runner.go#L60:3',
                                        '/runner.go#L66:3',
                                        '/mocks/initializer.go#L38:39',
                                        '/mocks/process_container.go#L73:30',
                                        '/mocks/process_container.go#L634:27',
                                        // ??? - too close to the one on the line above?
                                        // '/mocks/process_container.go#L635:29',
                                        '/mocks/process_container.go#L642:63',
                                        '/mocks/process_container.go#L651:84',
                                        '/mocks/process_container.go#L660:78',
                                        '/mocks/process_container.go#L669:32',
                                        '/mocks/process_container.go#L677:26',
                                        '/mocks/process_container.go#L682:75',
                                        '/mocks/process_container.go#L718:15',
                                        '/initializer_meta.go#L9:3',
                                        '/initializer_meta.go#L16:37',
                                        // ??? - too close to the one on the line above?
                                        // '/initializer_meta.go#L18:3',
                                        '/initializer_meta.go#L45:11',
                                    ].map(path => `/${repository}@${commit}/-/blob${path}`),
                                },
                            ],
                        },
                        {
                            path: '/watcher.go',
                            locations: [
                                {
                                    line: 166,
                                    token: 'logger',
                                    expectedHoverContains: 'struct field logger github.com/go-nacelle/log.Logger',
                                    expectedDefinition: `/${repository}@${commit}/-/blob/watcher.go#L39:2`,
                                    expectedReferences: [
                                        '/watcher.go#L39:2',
                                        '/watcher.go#L63:3',
                                        '/watcher.go#L105:7',
                                        '/watcher.go#L120:8',
                                        '/watcher.go#L125:8',
                                        '/watcher.go#L160:6',
                                        '/watcher.go#L166:5',
                                        '/watcher.go#L180:4',
                                        '/watcher.go#L202:5',
                                        '/watcher.go#L216:5',
                                        '/watcher_options.go#L16:37',
                                    ].map(path => `/${repository}@${commit}/-/blob${path}`),
                                },
                            ],
                        },
                    ],
                },
            ]

            await clearDumps(repository)
            await enableLSIF(driver, config)
            await uploadAndEnsureDump(driver, config, { repository, commit, root, filename })
            await testCodeIntel(driver, config, testCases)
        },
        60 * 1000
    )

    // TODO - test cross-repo
    // TODO - test closest commit
})
