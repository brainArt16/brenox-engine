import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 30,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const baseUrl = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const res = http.get(`${baseUrl}/health`);
  check(res, {
    'status is 200': (r) => r.status === 200,
    'body has status': (r) => r.body && r.body.includes('status'),
  });
  sleep(0.1);
}
