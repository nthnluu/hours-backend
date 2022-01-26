import './client_init'
import { getAuth, signInWithEmailAndPassword } from 'firebase/auth'
import { writeFile } from 'fs/promises'
import { join } from 'path'
import axios from 'axios'

export const BASE_DOMAIN = 'http://api.signmeup-load-test.uc.r.appspot.com'
const numUsers = 1
const auth = getAuth()

export interface TestUser {
    email: string
    cookie: string
}

export interface SetupData {
    testUsers: TestUser[]
}

export async function setup(): Promise<SetupData> {
    const testUsers: TestUser[] = []

    for (let i = 0; i < numUsers; i++) {
        const email = `tester-${i}@tester.com`
        const res = await signInWithEmailAndPassword(auth, email, 'san_diego_sucks_10!')
        console.log(`The user's display name is ${res.user.displayName}`)

        const idToken = await res.user.getIdToken(true)

        try {
            const sessionResponse = await axios.post(`${BASE_DOMAIN}/v1/users/session`, {
                token: idToken.toString(),
            })

            console.log(sessionResponse.data)

            const firstCookie = sessionResponse.headers['set-cookie'][0]
            const cookieText = firstCookie.substring(
                firstCookie.indexOf('=') + 1,
                firstCookie.indexOf(';'),
            )

            const meRes = await axios.get(`${BASE_DOMAIN}/v1/users/me`, {
                headers: {
                    Cookie: cookieText,
                },
            })

            console.log('me res is ' + JSON.stringify(meRes.data))

            testUsers.push({ email, cookie: cookieText })
        } catch (e) {
            console.log(e)
            return
        }
    }

    return { testUsers }
}

;(async () => {
    const testUsers = await setup()
    console.log('test users are ', testUsers)
    // const p = join(__dirname, './testUsers.json')
    // await writeFile(p, JSON.stringify(testUsers))
})()
