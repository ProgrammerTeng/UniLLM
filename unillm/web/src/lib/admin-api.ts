const API_BASE = "/api/admin";

async function adminFetch(path: string, options?: RequestInit) {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("token") : null;
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
  if (res.status === 403) {
    throw new Error("Admin access required");
  }
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export async function getGlobalStats() {
  return adminFetch("/stats");
}

export async function getUsers() {
  return adminFetch("/users");
}

export async function updateBalance(
  userId: number,
  delta: number,
  reason: string
) {
  return adminFetch("/users/balance", {
    method: "POST",
    body: JSON.stringify({ user_id: userId, delta, reason }),
  });
}

export async function getProviders() {
  return adminFetch("/providers");
}

export async function createProvider(name: string, baseUrl: string) {
  return adminFetch("/providers", {
    method: "POST",
    body: JSON.stringify({ name, base_url: baseUrl }),
  });
}

export async function toggleProvider(id: number, isActive: boolean) {
  return adminFetch("/providers/toggle", {
    method: "PUT",
    body: JSON.stringify({ id, is_active: isActive }),
  });
}

export async function getModels() {
  return adminFetch("/models");
}

export async function createModel(model: {
  public_name: string;
  provider_id: number;
  upstream_model: string;
  input_price_per_1m: number;
  output_price_per_1m: number;
  max_tokens: number;
}) {
  return adminFetch("/models", {
    method: "POST",
    body: JSON.stringify(model),
  });
}

export async function updateModel(updates: {
  id: number;
  input_price_per_1m?: number;
  output_price_per_1m?: number;
  is_active?: boolean;
  max_tokens?: number;
}) {
  return adminFetch("/models", {
    method: "PUT",
    body: JSON.stringify(updates),
  });
}

export async function getProviderKeys() {
  return adminFetch("/provider-keys");
}

export async function addProviderKey(
  providerId: number,
  keyValue: string,
  rpm: number
) {
  return adminFetch("/provider-keys", {
    method: "POST",
    body: JSON.stringify({ provider_id: providerId, key_value: keyValue, rpm }),
  });
}
