import m from 'mithril'

const API_BASE = '/api'

// Auth API
export const auth = {
  getCurrentUser: () =>
    m.request({
      method: 'GET',
      url: `${API_BASE}/auth/user`,
      withCredentials: true,
    }),
}

// Provider API
export const providers = {
  list: () =>
    m.request({
      method: 'GET',
      url: `${API_BASE}/providers`,
      withCredentials: true,
    }),

  create: (data) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/providers`,
      body: data,
      withCredentials: true,
    }),

  update: (id, data) =>
    m.request({
      method: 'PUT',
      url: `${API_BASE}/providers/${id}`,
      body: data,
      withCredentials: true,
    }),

  delete: (id) =>
    m.request({
      method: 'DELETE',
      url: `${API_BASE}/providers/${id}`,
      withCredentials: true,
    }),

  sync: (id) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/providers/${id}/sync`,
      withCredentials: true,
    }),

  analyze: (id) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/providers/${id}/analyze`,
      withCredentials: true,
    }),
}

// DNS Record API
export const records = {
  list: () =>
    m.request({
      method: 'GET',
      url: `${API_BASE}/records`,
      withCredentials: true,
    }),

  create: (data) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/records`,
      body: data,
      withCredentials: true,
    }),

  update: (id, data) =>
    m.request({
      method: 'PUT',
      url: `${API_BASE}/records/${id}`,
      body: data,
      withCredentials: true,
    }),

  delete: (id) =>
    m.request({
      method: 'DELETE',
      url: `${API_BASE}/records/${id}`,
      withCredentials: true,
    }),

  hide: (id) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/records/${id}/hide`,
      withCredentials: true,
    }),

  disable: (id) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/records/${id}/disable`,
      withCredentials: true,
    }),

  enable: (id) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/records/${id}/enable`,
      withCredentials: true,
    }),

  import: (data) =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/records/import`,
      body: data,
      withCredentials: true,
    }),

  reanalyze: () =>
    m.request({
      method: 'POST',
      url: `${API_BASE}/records/reanalyze`,
      withCredentials: true,
    }),
}

// Audit Log API
export const auditLogs = {
  list: (params) =>
    m.request({
      method: 'GET',
      url: `${API_BASE}/audit-logs`,
      params,
      withCredentials: true,
    }),
}
