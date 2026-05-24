import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (pathname.startsWith("/api/") || pathname.startsWith("/v1/") || pathname === "/backend/status" || pathname === "/backend/health" || pathname === "/backend/metrics") {
    // Map /backend/* to backend root paths
    const backendPath = pathname.startsWith("/backend/") ? pathname.replace("/backend", "") : pathname;
    const backendUrl = new URL(backendPath + request.nextUrl.search, BACKEND_URL);

    const headers = new Headers(request.headers);
    headers.set("host", new URL(BACKEND_URL).host);

    return NextResponse.rewrite(backendUrl, {
      request: { headers },
    });
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/api/:path*", "/v1/:path*", "/backend/:path*"],
};
