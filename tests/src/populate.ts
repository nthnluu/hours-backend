import './admin_init'
import { getAuth, UserImportRecord } from 'firebase-admin/auth'

const auth = getAuth()
const password = 'san_diego_sucks_10!'

const generateUsers = (numUsers: number): UserImportRecord[] => {
    const res: UserImportRecord[] = []

    for (let i = 0; i < numUsers; i++) {
        res.push({
            uid: `tester-${i}`,
            displayName: `Tester ${i}`,
            email: `tester-${i}@tester.com`,
        })
    }

    return res
}

async function driver(): Promise<void> {
    try {
        const numUsers = 1
        const res = await auth.importUsers(generateUsers(numUsers))
        if (res.errors.length > 0) {
            console.error(`Errors from importing: ${res.errors}. Exiting.`)
            return
        }
        console.info(
            `When creating ${numUsers}, ${res.successCount} successes, ${res.failureCount} failures`,
        )

        for (let i = 0; i < numUsers; i++) {
            await auth.updateUser(`tester-${i}`, { password })
            if (i % 100 === 0) {
                console.info(`Updated password for the ${i}th user`)
            }
        }

        console.info(`All done importing ${numUsers} users!`)
    } catch (e) {
        console.error('An unrecoverable error occurred: ' + e)
    }
}

driver()
