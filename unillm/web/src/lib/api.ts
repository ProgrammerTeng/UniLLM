const API_BASE = "/api";

async function fetchAPI(path: string, options?: RequestInit) {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options?.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401) {
    if (typeof window !== "undefined") {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    throw new Error("Unauthorized");
  }

  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export async function login(email: string, password: string) {
  const data = await fetchAPI("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  localStorage.setItem("token", data.token);
  return data;
}

export async function register(email: string, password: string, name: string) {
  const data = await fetchAPI("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, name }),
  });
  localStorage.setItem("token", data.token);
  return data;
}

export async function getMe() {
  return fetchAPI("/me");
}

export async function getUsageSummary() {
  return fetchAPI("/usage/summary");
}

export async function getUsageByModel() {
  return fetchAPI("/usage/by-model");
}

export async function getUsageDaily() {
  return fetchAPI("/usage/daily");
}

export async function getUsageRecent() {
  return fetchAPI("/usage/recent");
}

export async function getAPIKeys() {
  return fetchAPI("/keys");
}

export async function createAPIKey(name: string, scope: string = "full") {
  return fetchAPI("/keys", {
    method: "POST",
    body: JSON.stringify({ name, scope }),
  });
}

export async function deleteAPIKey(id: number) {
  return fetchAPI("/keys", {
    method: "DELETE",
    body: JSON.stringify({ id }),
  });
}

export function logout() {
  localStorage.removeItem("token");
  window.location.href = "/login";
}

export function isLoggedIn() {
  return typeof window !== "undefined" && !!localStorage.getItem("token");
}

export async function changePassword(oldPassword: string, newPassword: string) {
  return fetchAPI("/password", {
    method: "PUT",
    body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
  });
}

export async function getModelCatalog() {
  const res = await fetch(`${API_BASE}/models/catalog`);
  if (!res.ok) throw new Error("Failed to load model catalog");
  return res.json();
}
