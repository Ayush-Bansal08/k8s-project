import http from 'k6/http';
import { check, sleep } from 'k6';

// The traffic pattern: ramp virtual users up, hold, then ramp down
export const options = {
  stages: [
    { duration: '30s', target: 100 },  // ramp UP from 0 to 100 virtual users
    { duration: '1m',  target: 100 },  // HOLD at 100 users (the "storm")
    { duration: '20s', target: 0 },   // ramp DOWN to 0
  ],
};

// What each virtual user does, over and over
export default function () {
  const res = http.get('http://shop.localtest.me/');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);  // wait 1s, then hit it again
}
