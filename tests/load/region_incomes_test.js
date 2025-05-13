import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';


const errorRate = new Rate('errors');

export const options = {
    stages: [
        { duration: '15s', target: 100 },    
        { duration: '15s', target: 500 },    
        { duration: '15s', target: 1000 },   
        { duration: '15s', target: 0 },    
    ],
    
    thresholds: {
        http_req_duration: ['p(95)<500'],  
        errors: ['rate<0.1'],              
        'http_req_failed': ['rate<0.1'],   
    },
};

function getRandomParams() {
    return {
        regionid: Math.floor(Math.random() * 79) + 1,
    };
}

export default function() {
    const params = getRandomParams();
    
    const url = `http://localhost:8080/api/v1/regionincomes?regionid=${params.regionid}`;
    
    const res = http.get(url, {
        tags: {
            region: params.regionid,
        },
    });

    const checkResult = check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
        'has valid JSON': (r) => {
            try {
                JSON.parse(r.body);
                return true;
            } catch (e) {
                return false;
            }
        },
    });

    errorRate.add(!checkResult);

    sleep(1);
}