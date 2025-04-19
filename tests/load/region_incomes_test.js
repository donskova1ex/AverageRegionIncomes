import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
    
    stages: [
        { duration: '15s', target: 100 },    
        { duration: '15s', target: 500 },    
        { duration: '15s', target: 1000 },   
        { duration: '15s', target: 10 },     
    ],
    
    
    thresholds: {
        http_req_duration: ['p(95)<500'],  
        errors: ['rate<0.1'],              
        'http_req_failed': ['rate<0.1'],   
    },
};

function getRandomParams() {
    const params = {
        regionid: Math.floor(Math.random() * 16) + 1
    };
    
    if (Math.random() > 0.5) {
        params.year = Math.floor(Math.random() * 2) + 2023;
        
        if (Math.random() > 0.5) {
            params.quarter = Math.floor(Math.random() * 4) + 1;
        }
    }
    
    return params;
}

function buildQueryString(params) {
    return Object.entries(params)
        .map(([key, value]) => `${key}=${value}`)
        .join('&');
}

export default function() {
    const params = getRandomParams();
    const queryString = buildQueryString(params);
    const url = `http://localhost:8080/api/v1/regionincomes?${queryString}`;
    
    const res = http.get(url, {
        headers: {
            'Accept': 'application/json'
        }
    });

    const checkResult = check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
        'has valid JSON': (r) => {
            try {
                JSON.parse(r.body);
                return true;
            } catch (e) {
                console.log(`Invalid JSON response: ${r.body}`);
                return false;
            }
        },
    });

    if (res.status !== 200) {
        console.log(`Failed request: ${url}`);
        console.log(`Status: ${res.status}`);
        console.log(`Response: ${res.body}`);
    }

    errorRate.add(!checkResult);

    sleep(1);
}