const API_BASE = '/api'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null
  const res = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  })
  const data = await res.json()
  if (data.code !== 0) throw new Error(data.msg)
  return data.data
}

export const api = {
  get: <T>(url: string) => request<T>(url),
  post: <T>(url: string, body?: unknown) => request<T>(url, { method: 'POST', body: JSON.stringify(body) }),
  put: <T>(url: string, body?: unknown) => request<T>(url, { method: 'PUT', body: JSON.stringify(body) }),
  del: <T>(url: string, body?: unknown) => request<T>(url, { method: 'DELETE', body: body ? JSON.stringify(body) : undefined }),
  upload: async <T>(url: string, file: File) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null
    const fd = new FormData()
    fd.append('file', file)
    const res = await fetch(`${API_BASE}${url}`, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: fd,
    })
    const data = await res.json()
    if (data.code !== 0) throw new Error(data.msg)
    return data.data as T
  },
}
