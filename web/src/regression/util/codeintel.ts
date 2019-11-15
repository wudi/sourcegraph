import * as child_process from 'mz/child_process'
import * as path from 'path'
import got from 'got'
import { Config } from '../../../../shared/src/e2e/config'
import { Driver } from '../../../../shared/src/e2e/driver'

export interface CodeNavTestCase {
    repoRev: string
    files: {
        path: string
        locations: {
            line: number
            token: string
            expectedHoverContains: string
            expectedDefinition: string | string[]
            expectedReferences?: string[]
        }[]
    }[]
}

export async function testCodeIntel(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    testCases: CodeNavTestCase[]
): Promise<void> {
    const getTooltip = () =>
        driver.page.evaluate(() => (document.querySelector('.e2e-tooltip-content') as HTMLElement).innerText)
    const collectLinks = (selector: string) =>
        driver.page.evaluate(selector => {
            const links: string[] = []
            document.querySelectorAll<HTMLElement>(selector).forEach(e => {
                e.querySelectorAll<HTMLElement>('a[href]').forEach(a => {
                    const href = a.getAttribute('href')
                    if (href) {
                        links.push(href)
                    }
                })
            })
            return links
        }, selector)
    const clickOnEmptyPartOfCodeView = () => driver.page.$x('//*[contains(@class, "e2e-blob")]//tr[1]//*[text() = ""]')
    const findTokenElement = async (line: number, token: string) => {
        const xpathQuery = `//*[contains(@class, "e2e-blob")]//tr[${line}]//*[normalize-space(text()) = ${JSON.stringify(
            token
        )}]`
        return {
            tokenEl: await driver.page.$x(xpathQuery),
            xpathQuery,
        }
    }
    const normalizeWhitespace = (s: string) => s.replace(/\s+/g, ' ')
    const waitForHover = async (expectedHoverContains: string) => {
        await driver.page.waitForSelector('.e2e-tooltip-go-to-definition')
        await driver.page.waitForSelector('.e2e-tooltip-content')
        expect(normalizeWhitespace(await getTooltip())).toContain(normalizeWhitespace(expectedHoverContains))
    }

    for (const { repoRev, files } of testCases) {
        for (const { path, locations } of files) {
            await driver.page.goto(config.sourcegraphBaseUrl + `/${repoRev}/-/blob${path}`)
            await driver.page.waitForSelector('.e2e-blob')
            for (const { line, token, expectedHoverContains, expectedDefinition, expectedReferences } of locations) {
                const { tokenEl, xpathQuery } = await findTokenElement(line, token)
                if (tokenEl.length === 0) {
                    throw new Error(
                        `did not find token ${JSON.stringify(token)} on page. XPath query was: ${xpathQuery}`
                    )
                }

                // Check hover and click
                await tokenEl[0].hover() // TODO - race condition when LSIF enabled?
                await waitForHover(expectedHoverContains)
                const { tokenEl: emptyTokenEl } = await findTokenElement(line, '')
                await emptyTokenEl[0].hover()
                await driver.page.waitForFunction(
                    () => document.querySelectorAll('.e2e-tooltip-go-to-definition').length === 0
                )
                await tokenEl[0].click()
                await waitForHover(expectedHoverContains)

                // Find-references
                if (expectedReferences) {
                    await (await driver.findElementWithText('Find references')).click()
                    await driver.page.waitForSelector('.e2e-search-result')
                    const refLinks = await collectLinks('.e2e-search-result')
                    for (const expectedReference of expectedReferences) {
                        if (!refLinks.includes(expectedReference)) {
                            console.log({ refLinks, expectedReference })
                        }
                        expect(refLinks.includes(expectedReference)).toBeTruthy()
                    }
                    await clickOnEmptyPartOfCodeView()
                }

                // Go-to-definition
                await (await driver.findElementWithText('Go to definition')).click()
                if (Array.isArray(expectedDefinition)) {
                    await driver.page.waitForSelector('.e2e-search-result')
                    const defLinks = await collectLinks('.e2e-search-result')
                    expect(expectedDefinition.every(l => defLinks.includes(l))).toBeTruthy()
                } else {
                    await driver.page.waitForFunction(
                        defURL => document.location.href.endsWith(defURL),
                        { timeout: 20000 },
                        expectedDefinition
                    )
                    await driver.page.goBack()
                }

                await driver.page.keyboard.press('Escape')
            }
        }
    }
}

//
//

export async function clearDumps(repository: string): Promise<void> {
    // TODO - do through graphql
    const url = new URL(`http://localhost:3186/dumps/${encodeURIComponent(repository)}`)
    const resp = await got.get(url.href)
    const body: { dumps: { id: string }[] } = JSON.parse(resp.body)
    await Promise.all(
        body.dumps.map(dump => {
            const url = new URL(`http://localhost:3186/dumps/${encodeURIComponent(repository)}/${dump.id}`)
            return got.delete(url.href)
        })
    )
}

export async function enableLSIF(driver: Driver, config: Pick<Config, 'sourcegraphBaseUrl'>): Promise<void> {
    await driver.page.goto(`${config.sourcegraphBaseUrl}/site-admin/global-settings`)
    const globalSettings = '{"codeIntel.lsif": true}'
    await driver.replaceText({
        selector: '.monaco-editor',
        newText: globalSettings,
        selectMethod: 'keyboard',
        enterTextMethod: 'type',
    })
    await (
        await driver.findElementWithText('Save changes', {
            selector: 'button',
            wait: { timeout: 500 },
        })
    ).click()
}

export async function uploadAndEnsureDump(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    {
        repository,
        commit,
        root,
        filename,
    }: {
        repository: string
        commit: string
        root: string
        filename: string
    }
): Promise<void> {
    const jobUrl = await uploadDump({
        endpoint: config.sourcegraphBaseUrl,
        repository,
        commit,
        root,
        filename,
    })

    await ensureJobCompleted(driver, jobUrl)
    await updateTips()
    await driver.page.goto(`${config.sourcegraphBaseUrl}/${repository}/-/settings/code-intelligence`)

    expect(await (await driver.page.waitForSelector('.e2e-dump-commit')).evaluate(elem => elem.textContent)).toEqual(
        commit.substr(0, 7)
    )

    expect(await (await driver.page.waitForSelector('.e2e-dump-path')).evaluate(elem => elem.textContent)).toEqual(root)
}

async function uploadDump({
    endpoint,
    repository,
    commit,
    root,
    filename,
}: {
    endpoint: string
    repository: string
    commit: string
    root: string
    filename: string
}): Promise<string> {
    let out!: Buffer
    try {
        // TODO - ensure binary is installed
        ;[out] = await child_process.exec(
            [
                `src-cli -endpoint ${endpoint}`,
                'lsif upload',
                `-repo ${repository}`,
                `-commit ${commit}`,
                `-root ${root}`,
                `-file ${filename}`,
            ].join(' '),

            { cwd: path.join(__dirname, '..') }
        )
    } catch (error) {
        if (error && error.stdout) {
            throw new Error(`Failed to upload LSIF dump: ${error.stdout}`)
        }
        throw error
    }

    const match = out.toString().match(/To check the status, visit (.+).\n$/)
    if (!match) {
        throw new Error(`Unexpected output from Sourcegraph cli: ${out.toString()}`)
    }

    return match[1]
}

const pendingJobStateMessages = ['Job is queued.', 'Job is currently being processed...']

async function ensureJobCompleted(driver: Driver, url: string): Promise<void> {
    await driver.page.goto(url)
    while (true) {
        const text = await (await driver.page.waitForSelector('.e2e-job-state')).evaluate(elem => elem.textContent)
        if (!pendingJobStateMessages.includes(text || '')) {
            break
        }

        await driver.page.reload()
    }

    const text = await (await driver.page.waitForSelector('.e2e-job-state')).evaluate(elem => elem.textContent)
    expect(text).toEqual('Job completed successfully.')
}

async function updateTips(): Promise<void> {
    // TODO - do through graphql
    const resp = await got.post(new URL('http://localhost:3186/jobs?blocking=true').href, {
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ name: 'update-tips' }),
    })
    expect(resp.statusCode).toEqual(200)
}
