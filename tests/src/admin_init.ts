import { initializeApp, applicationDefault } from 'firebase-admin/app'

// GOOGLE_APPLICATION_CREDENTIALS is set to staging-firebase-config.json
export const app = initializeApp({
    credential: applicationDefault(),
})
