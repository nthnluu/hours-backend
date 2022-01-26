import { check, sleep } from 'k6'
import http from 'k6/http'

export const options = {
    stages: [
        { duration: '30s', target: 30 },
        { duration: '5s', target: 300 },
    ],
}

export default function () {
    const jar = http.cookieJar()
    jar.set(
        'https://api.hours.luu.dev/v1/queues/tDPFJhtlDLobeV5LbDMN/ticket',
        'hours-session',
        'haha',
    )

    const res = http.post(
        'https://api.hours.luu.dev/v1/queues/tDPFJhtlDLobeV5LbDMN/ticket',
        JSON.stringify({
            description: 'Hello from K6',
        }),
    )

    check(res, { 'queue signup was successful': (r) => r.status === 200 })

    sleep(1)
}
