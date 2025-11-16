import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    scenarios: {
        demo_test: {
            executor: 'constant-arrival-rate',
            rate: 500,
            timeUnit: '1s',
            duration: '3m',
            preAllocatedVUs: 100,
            maxVUs: 500,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<300'],   // SLI responce time    ≤ 300 мс
        checks: ['rate>0.999'],             // SLI success          ≥ 99.9%
    },
};

const baseUrl = 'http://app:8080';

// Generate team with members
function createTeamData() {
    const teamName = `team-${randomString(10)}`;
    const membersCount = randomIntBetween(5, 10); // Random members amount
    const members = [];
    for (let i = 0; i < membersCount; i++) {
        members.push({
            user_id: `u-${randomString(10)}`,
            username: `User-${randomString(5)}`,
            is_active: true,
        });
    }
    return { teamName, members };
}

export default function () {
    const teamData = createTeamData();

    // --- 1. Team creating ---
    const teamRes = http.post(`${baseUrl}/team/add`, JSON.stringify({
        team_name: teamData.teamName,
        members: teamData.members,
    }), { headers: { 'Content-Type': 'application/json' } });

    if (!check(teamRes, { 'team created': (r) => r.status === 201 })) {
        console.error('Failed to create team:', teamRes.status, teamRes.body);
        return;
    }

    // --- 2. Getting team ---
    const getTeamRes = http.get(`${baseUrl}/team/get?team_name=${teamData.teamName}`);
    if (!check(getTeamRes, { 'team fetched': (r) => r.status === 200 })) {
        console.error('Failed to fetch team:', getTeamRes.status, getTeamRes.body);
        return;
    }

    const createdTeam = getTeamRes.json() || {};
    const members = createdTeam.members || [];

    // --- 3. Creating PR for each team member ---
    const prs = [];
    members.forEach((member) => {
        const prId = `pr-${randomString(10)}`;
        const prName = `Feature-${randomString(10)}`;
        const prRes = http.post(`${baseUrl}/pullRequest/create`, JSON.stringify({
            pull_request_id: prId,
            pull_request_name: prName,
            author_id: member.user_id
        }), { headers: { 'Content-Type': 'application/json' } });

        if (check(prRes, { 'PR created': (r) => r.status === 201 })) {
            prs.push({ prId, reviewers: prRes.json('pr.assigned_reviewers') || [] });
        } else {
            console.error(`Failed to create PR for ${member.username}:`, prRes.status, prRes.body);
        }
    });

    // --- 5. For each member get PRs where he is reviewer ---
    members.forEach((member) => {
        const res = http.get(`${baseUrl}/users/getReview?user_id=${member.user_id}`);
        check(res, { 'getReview success': (r) => r.status === 200 });
    });

    sleep(1);
}
