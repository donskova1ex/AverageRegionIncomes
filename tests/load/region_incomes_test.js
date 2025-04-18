import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';


const errorRate = new Rate('errors');

export const options = {
    stages: [
        { duration: '1m', target: 10 },    
        { duration: '1m', target: 50 },    
        { duration: '1m', target: 100 },   
        { duration: '1m', target: 0 },    
    ],
    
    thresholds: {
        http_req_duration: ['p(95)<500'],  
        errors: ['rate<0.1'],              
        'http_req_failed': ['rate<0.1'],   
    },
};

function getRandomParams() {
    return {
        regionid: Math.floor(Math.random() * 85) + 1,
        year: Math.floor(Math.random() * 3) + 2022,
        quarter: Math.floor(Math.random() * 4) + 1,
    };
}

export default function() {
    const params = getRandomParams();
    
    const res = http.get('http://localhost:8080/api/v1/regionincomes', {
        params: params,
        tags: {
            region: params.regionid,
            year: params.year,
            quarter: params.quarter,
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