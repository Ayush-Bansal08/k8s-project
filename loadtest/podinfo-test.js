import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 10,             // 10 virtual users
  duration: '10m',     // run for 10 minutes (covers the whole canary demo)
};

export default function () {
  http.get('http://localhost:9900/');   // hits podinfo via the port-forward
  sleep(0.2);
}
