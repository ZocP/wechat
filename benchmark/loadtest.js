import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ============================================================
//  接机管理系统 (Pickup Service) — k6 压力测试脚本
//  测试目标: Gin + GORM + MySQL 全链路核心 API
// ============================================================

const BASE_URL = 'http://localhost:9090/api/v1';

const TOKENS = JSON.parse(open('./tokens.json').replace(/^\uFEFF/, '').trim());

const JSON_HEADERS = {
  headers: {
    'Content-Type': 'application/json',
  },
};

// ─── 自定义指标 ──────────────────────────────────────────
const healthCheckDuration   = new Trend('health_check_duration', true);
const createRegDuration     = new Trend('create_registration_duration', true);
const listRegDuration       = new Trend('list_registrations_duration', true);
const getRegDuration        = new Trend('get_registration_duration', true);
const updateRegDuration     = new Trend('update_registration_duration', true);
const deleteRegDuration     = new Trend('delete_registration_duration', true);
const createOrderDuration   = new Trend('create_order_duration', true);
const listOrdersDuration    = new Trend('list_orders_duration', true);
const listNoticesDuration   = new Trend('list_notices_duration', true);
const createNoticeDuration  = new Trend('create_notice_duration', true);
const errorRate             = new Rate('error_rate');
const successfulRequests    = new Counter('successful_requests');

// ─── 测试配置：多阶段递增压力 ─────────────────────────────
export const options = {
  stages: [
    { duration: '15s', target: 20 },   // 预热: 0 → 20 VU
    { duration: '30s', target: 50 },   // 爬坡: → 50 VU
    { duration: '60s', target: 100 },  // 峰值负载: 100 VU 持续 60s
    { duration: '30s', target: 100 },  // 稳态: 保持 100 VU
    { duration: '15s', target: 0 },    // 冷却: → 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],  // P95 < 500ms, P99 < 1s
    error_rate: ['rate<0.05'],                        // 错误率 < 5%
    http_req_failed: ['rate<0.05'],
  },
};

// ─── 辅助函数 ─────────────────────────────────────────────
let regCounter = 0;

function uniqueId() {
  return `${__VU}-${__ITER}-${Date.now()}`;
}

function makeRegistrationPayload() {
  const uid = uniqueId();
  return JSON.stringify({
    name: `压测用户_${uid}`,
    phone: `138${String(Math.floor(Math.random() * 100000000)).padStart(8, '0')}`,
    wechat_id: `wx_${uid}`,
    flight_no: `CA${1000 + Math.floor(Math.random() * 9000)}`,
    arrival_date: '2026-03-01',
    arrival_time: '14:30',
    departure_city: '北京',
    companions: Math.floor(Math.random() * 4),
    luggage_count: Math.floor(Math.random() * 6) + 1,
    pickup_method: ['group', 'private', 'shuttle'][Math.floor(Math.random() * 3)],
    notes: `k6压测数据 ${uid}`,
  });
}

function makeNoticePayload() {
  const uid = uniqueId();
  const now = new Date();
  const visibleFrom = now.toISOString();
  const visibleTo = new Date(now.getTime() + 86400000).toISOString();
  return JSON.stringify({
    flight_no: `CA${1000 + Math.floor(Math.random() * 9000)}`,
    terminal: 'T3',
    pickup_batch: `BATCH_${uid}`,
    arrival_airport: '首都国际机场',
    meeting_point: 'T3到达层3号门',
    guide_text: '请在到达层3号门集合',
    map_url: 'https://example.com/map',
    contact_name: '压测调度员',
    contact_phone: '13900000000',
    visible_from: visibleFrom,
    visible_to: visibleTo,
  });
}

function authHeaders(token) {
  return {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  };
}

function pickRandomToken() {
  return TOKENS[Math.floor(Math.random() * TOKENS.length)];
}

// ─── 主测试函数 ──────────────────────────────────────────
export default function () {
  const currentToken = pickRandomToken();
  const currentAuthHeaders = authHeaders(currentToken);

  // ===== 1. 健康检查 =====
  group('健康检查 - GET /health', () => {
    const res = http.get(`${BASE_URL}/health`);
    healthCheckDuration.add(res.timings.duration);
    const ok = check(res, {
      'health status 200': (r) => r.status === 200,
      'health body ok':   (r) => r.json('status') === 'ok',
    });
    errorRate.add(!ok);
    if (ok) successfulRequests.add(1);
  });

  // ===== 2. 报名 CRUD 全流程 =====
  let registrationId = null;

  group('创建报名 - POST /registrations', () => {
    const res = http.post(`${BASE_URL}/registrations`, makeRegistrationPayload(), currentAuthHeaders);
    createRegDuration.add(res.timings.duration);
    const ok = check(res, {
      'create reg status 201': (r) => r.status === 201,
      'create reg code 0':     (r) => r.json('code') === 0,
    });
    errorRate.add(!ok);
    if (ok) {
      successfulRequests.add(1);
      try { registrationId = res.json('data.id') || res.json('data.ID'); } catch(e) {}
    }
  });

  group('查询报名列表 - GET /registrations', () => {
    const res = http.get(`${BASE_URL}/registrations?page=1&page_size=20`, currentAuthHeaders);
    listRegDuration.add(res.timings.duration);
    const ok = check(res, {
      'list reg status 200': (r) => r.status === 200,
      'list reg code 0':     (r) => r.json('code') === 0,
    });
    errorRate.add(!ok);
    if (ok) successfulRequests.add(1);
  });

  if (registrationId) {
    group('查询报名详情 - GET /registrations/:id', () => {
      const res = http.get(`${BASE_URL}/registrations/${registrationId}`, currentAuthHeaders);
      getRegDuration.add(res.timings.duration);
      const ok = check(res, {
        'get reg status 200': (r) => r.status === 200,
      });
      errorRate.add(!ok);
      if (ok) successfulRequests.add(1);
    });

    group('更新报名 - PUT /registrations/:id', () => {
      const updatePayload = JSON.stringify({
        notes: `更新于 ${Date.now()}`,
        companions: Math.floor(Math.random() * 3) + 1,
      });
      const res = http.put(`${BASE_URL}/registrations/${registrationId}`, updatePayload, currentAuthHeaders);
      updateRegDuration.add(res.timings.duration);
      const ok = check(res, {
        'update reg status 200': (r) => r.status === 200,
      });
      errorRate.add(!ok);
      if (ok) successfulRequests.add(1);
    });

    // ===== 3. 订单 =====
    group('创建订单 - POST /orders', () => {
      const orderPayload = JSON.stringify({
        registration_id: registrationId,
        price_total: Math.floor(Math.random() * 500 + 100) * 100,
        currency: 'CNY',
      });
      const res = http.post(`${BASE_URL}/orders`, orderPayload, currentAuthHeaders);
      createOrderDuration.add(res.timings.duration);
      const ok = check(res, {
        'create order status 201': (r) => r.status === 201,
        'create order code 0':     (r) => r.json('code') === 0,
      });
      errorRate.add(!ok);
      if (ok) successfulRequests.add(1);
    });

  }

  group('查询订单列表 - GET /orders', () => {
    const res = http.get(`${BASE_URL}/orders?page=1&page_size=20`, currentAuthHeaders);
    listOrdersDuration.add(res.timings.duration);
    const ok = check(res, {
      'list orders status 200': (r) => r.status === 200,
    });
    errorRate.add(!ok);
    if (ok) successfulRequests.add(1);
  });

  // ===== 4. 消息通知（公开接口） =====
  group('查询消息列表 - GET /notices', () => {
    const res = http.get(`${BASE_URL}/notices?page=1&page_size=20`, JSON_HEADERS);
    listNoticesDuration.add(res.timings.duration);
    const ok = check(res, {
      'list notices status 200': (r) => r.status === 200,
      'list notices code 0':     (r) => r.json('code') === 0,
    });
    errorRate.add(!ok);
    if (ok) successfulRequests.add(1);
  });

  group('创建消息(管理端) - POST /admin/notices', () => {
    const res = http.post(`${BASE_URL}/admin/notices`, makeNoticePayload(), currentAuthHeaders);
    createNoticeDuration.add(res.timings.duration);
    const ok = check(res, {
      'create notice status 201': (r) => r.status === 201,
    });
    errorRate.add(!ok);
    if (ok) successfulRequests.add(1);
  });

  sleep(0.3); // 模拟用户思考间隔
}

// ─── 测试结束汇总 ────────────────────────────────────────
export function handleSummary(data) {
  // 输出 JSON 结果文件
  return {
    'benchmark/results.json': JSON.stringify(data, null, 2),
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, opts) {
  // k6 内置 summary 由运行时自动输出, 这里返回空字符即可
  return '';
}
