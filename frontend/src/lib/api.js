// Simple API helper that reads VITE_API_URL (set in .env or docker-compose)
const API = import.meta.env.VITE_API_URL || '';

export async function createUser(payload) {
  const res = await fetch(`${API}/users`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return res.json();
}

export async function listUsers() {
  const res = await fetch(`${API}/users`);
  return res.json();
}

export async function createConsumption(userId, payload) {
  const res = await fetch(`${API}/users/${userId}/consumptions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return res.json();
}

export async function listConsumptions(q = {}) {
  const params = new URLSearchParams(q).toString();
  const res = await fetch(`${API}/consumptions${params ? '?' + params : ''}`);
  return res.json();
}

export default {
  createUser,
  listUsers,
  createConsumption,
  listConsumptions,
};
