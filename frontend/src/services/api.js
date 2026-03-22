import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000/api';

// Create an Axios instance with base URL
const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor to inject JWT token into all requests
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Interceptor to handle 401 Unauthorized globally (e.g. expired token)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // Clear storage and redirect to login if not already there
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// ──────────────────────────────────────────────────────────────────────
// Auth Services
// ──────────────────────────────────────────────────────────────────────

export const login = async (username, password) => {
  const { data } = await api.post('/login', { username, password });
  if (data.token) {
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
  }
  return data;
};

export const logout = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
};

// ──────────────────────────────────────────────────────────────────────
// Directory & Common Services
// ──────────────────────────────────────────────────────────────────────

export const getDirectory = async () => {
  const { data } = await api.get('/directory');
  return data;
};

// ──────────────────────────────────────────────────────────────────────
// Level 2 Doctor Services
// ──────────────────────────────────────────────────────────────────────

export const checkNotifications = async () => {
  const { data } = await api.get('/notifications');
  return data;
};

export const markNotificationRead = async (id) => {
  const { data } = await api.patch(`/notifications/${id}/read`);
  return data;
};

export const suggestDepartment = async (payload) => {
  const { data } = await api.post('/referrals/suggest', payload);
  return data;
};

export const createReferral = async (payload) => {
  const { data } = await api.post('/referrals', payload);
  return data;
};

export const uploadAttachments = async (referralID, files) => {
  const formData = new FormData();
  for (let i = 0; i < files.length; i++) {
    formData.append('attachments', files[i]);
  }
  const { data } = await api.post(`/referrals/${referralID}/attachments`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return data;
};

export const getAttachmentUrl = (id) => `${API_URL}/attachments/${id}`;

// ──────────────────────────────────────────────────────────────────────
// CHU Doctor Services
// ──────────────────────────────────────────────────────────────────────

export const getQueue = async () => {
  const { data } = await api.get('/queue');
  return data;
};

export const getReferralDetails = async (id) => {
  const { data } = await api.get(`/referrals/${id}`);
  return data;
};

export const scheduleReferral = async (id, appointmentDate) => {
  const { data } = await api.patch(`/referrals/${id}/schedule`, { appointment_date: appointmentDate });
  return data;
};

export const redirectReferral = async (id, newDepartmentID, reason) => {
  const { data } = await api.patch(`/referrals/${id}/redirect`, {
    new_department_id: newDepartmentID,
    reason: reason,
  });
  return data;
};

export const denyReferral = async (id, reason) => {
  const { data } = await api.patch(`/referrals/${id}/deny`, { reason });
  return data;
};

// ── Admin Endpoints ──
export const getAdminStats = async () => {
  const { data } = await api.get('/admin/stats');
  return data;
};

export const getUsers = async () => {
  const { data } = await api.get('/admin/users');
  return data;
};

export const createUser = async (payload) => {
  const { data } = await api.post('/admin/users', payload);
  return data;
};

export const deleteUser = async (id) => {
  const { data } = await api.delete(`/admin/users/${id}`);
  return data;
};

export const getAdminDepartments = async () => {
  const { data } = await api.get('/admin/departments');
  return data;
};

export const createDepartment = async (payload) => {
  const { data } = await api.post('/admin/departments', payload);
  return data;
};

export const updateDepartment = async (id, payload) => {
  const { data } = await api.patch(`/admin/departments/${id}`, payload);
  return data;
};

export const deleteDepartment = async (id) => {
  const { data } = await api.delete(`/admin/departments/${id}`);
  return data;
};

// ── Analyst Endpoints ──
export const getAnalystStats = async () => {
  const { data } = await api.get('/analyst/stats/departments');
  return data;
};

export const getAnalystDoctorStats = async () => {
  const { data } = await api.get('/analyst/stats/doctors');
  return data;
};

// ── History & Rescheduling ──
export const getHistory = async () => {
  const { data } = await api.get('/history');
  return data;
};

export const rescheduleReferral = async (id, appointmentDate) => {
  const { data } = await api.patch(`/referrals/${id}/reschedule`, { appointment_date: appointmentDate });
  return data;
};

export const cancelReferral = async (id) => {
  const { data } = await api.patch(`/referrals/${id}/cancel`);
  return data;
};

export default api;
