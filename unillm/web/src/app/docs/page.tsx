"use client";

import { useState } from "react";

const SDK_LANGUAGES = ["Python", "JavaScript", "Go", "Java", "Ruby", "PHP", "curl"] as const;
type SdkLanguage = (typeof SDK_LANGUAGES)[number];

const SDK_EXAMPLES: Record<SdkLanguage, { lang: string; code: string }> = {
  Python: {
    lang: "python",
    code: `from openai import OpenAI

client = OpenAI(
    api_key="sk-your-key",
    base_url="https://your-domain/v1"
)

response = client.chat.completions.create(
    model="deepseek-v3",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)`,
  },
  JavaScript: {
    lang: "typescript",
    code: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-your-key",
  baseURL: "https://your-domain/v1",
});

const response = await client.chat.completions.create({
  model: "claude-haiku",
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(response.choices[0].message.content);`,
  },
  Go: {
    lang: "go",
    code: `package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    body, _ := json.Marshal(map[string]interface{}{
        "model": "deepseek-v3",
        "messages": []map[string]string{
            {"role": "user", "content": "Hello!"},
        },
    })

    req, _ := http.NewRequest("POST",
        "https://your-domain/v1/chat/completions",
        bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer sk-your-key")
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    data, _ := io.ReadAll(resp.Body)
    fmt.Println(string(data))
}`,
  },
  Java: {
    lang: "java",
    code: `import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class UniLLMExample {
    public static void main(String[] args) throws Exception {
        String json = """
            {
              "model": "deepseek-v3",
              "messages": [{"role": "user", "content": "Hello!"}]
            }
            """;

        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create("https://your-domain/v1/chat/completions"))
            .header("Authorization", "Bearer sk-your-key")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(json))
            .build();

        HttpResponse<String> response = HttpClient.newHttpClient()
            .send(request, HttpResponse.BodyHandlers.ofString());
        System.out.println(response.body());
    }
}`,
  },
  Ruby: {
    lang: "ruby",
    code: `require 'net/http'
require 'json'
require 'uri'

uri = URI("https://your-domain/v1/chat/completions")
http = Net::HTTP.new(uri.host, uri.port)
http.use_ssl = true

request = Net::HTTP::Post.new(uri)
request["Authorization"] = "Bearer sk-your-key"
request["Content-Type"] = "application/json"
request.body = {
  model: "deepseek-v3",
  messages: [{ role: "user", content: "Hello!" }]
}.to_json

response = http.request(request)
result = JSON.parse(response.body)
puts result["choices"][0]["message"]["content"]`,
  },
  PHP: {
    lang: "php",
    code: `<?php
$ch = curl_init("https://your-domain/v1/chat/completions");

$payload = json_encode([
    "model" => "deepseek-v3",
    "messages" => [
        ["role" => "user", "content" => "Hello!"]
    ]
]);

curl_setopt_array($ch, [
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_POST => true,
    CURLOPT_POSTFIELDS => $payload,
    CURLOPT_HTTPHEADER => [
        "Authorization: Bearer sk-your-key",
        "Content-Type: application/json"
    ]
]);

$response = curl_exec($ch);
curl_close($ch);

$result = json_decode($response, true);
echo $result["choices"][0]["message"]["content"];`,
  },
  curl: {
    lang: "bash",
    code: `curl https://your-domain/v1/chat/completions \\
  -H "Authorization: Bearer sk-your-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gemini-flash",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`,
  },
};

const SECTIONS = [
  {
    id: "quickstart",
    title: "Quick Start",
    content: `## Getting Started

1. **Register** an account at the UniLLM dashboard
2. **Create an API key** from the API Keys tab
3. **Use your key** with any OpenAI-compatible SDK

UniLLM provides a single API endpoint that routes to multiple AI providers (OpenAI, Anthropic Claude, Google Gemini, DeepSeek, etc.) using the OpenAI-compatible format.`,
  },
  {
    id: "auth",
    title: "Authentication",
    content: `## Authentication

All API requests require an API key in the Authorization header:

\`\`\`
Authorization: Bearer sk-your-api-key-here
\`\`\`

API keys can be created and managed from the dashboard.`,
  },
  {
    id: "chat",
    title: "Chat Completions",
    content: `## POST /v1/chat/completions

Create a chat completion. Supports streaming and non-streaming modes.

### Request Body

\`\`\`json
{
  "model": "deepseek-v3",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false,
  "temperature": 0.7,
  "max_tokens": 1024
}
\`\`\`

### Response

\`\`\`json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "model": "deepseek-v3",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "Hello! How can I help?"},
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 8,
    "total_tokens": 28
  }
}
\`\`\``,
  },
  {
    id: "models",
    title: "List Models",
    content: `## GET /v1/models

Returns available models.

### Response

\`\`\`json
{
  "object": "list",
  "data": [
    {"id": "deepseek-v3", "object": "model", "owned_by": "unillm"},
    {"id": "claude-haiku", "object": "model", "owned_by": "unillm"},
    {"id": "gemini-flash", "object": "model", "owned_by": "unillm"}
  ]
}
\`\`\``,
  },
  {
    id: "streaming",
    title: "Streaming",
    content: `## Streaming

Set \`"stream": true\` to receive Server-Sent Events (SSE).

\`\`\`
data: {"id":"xxx","choices":[{"delta":{"content":"Hello"},"index":0}]}

data: {"id":"xxx","choices":[{"delta":{"content":" world"},"index":0}]}

data: [DONE]
\`\`\`

### Example (curl)

\`\`\`bash
curl -N https://your-domain/v1/chat/completions \\
  -H "Authorization: Bearer sk-xxx" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"deepseek-v3","messages":[{"role":"user","content":"Hi"}],"stream":true}'
\`\`\`

### Parsing SSE in Python

\`\`\`python
from openai import OpenAI

client = OpenAI(api_key="sk-your-key", base_url="https://your-domain/v1")

stream = client.chat.completions.create(
    model="deepseek-v3",
    messages=[{"role": "user", "content": "Hello!"}],
    stream=True
)

for chunk in stream:
    delta = chunk.choices[0].delta
    if delta.content:
        print(delta.content, end="", flush=True)
print()
\`\`\`

### Parsing SSE in JavaScript

\`\`\`typescript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-your-key",
  baseURL: "https://your-domain/v1",
});

const stream = await client.chat.completions.create({
  model: "deepseek-v3",
  messages: [{ role: "user", content: "Hello!" }],
  stream: true,
});

for await (const chunk of stream) {
  const content = chunk.choices[0]?.delta?.content;
  if (content) process.stdout.write(content);
}
\`\`\`

### Raw SSE Parsing (fetch)

\`\`\`typescript
const response = await fetch("https://your-domain/v1/chat/completions", {
  method: "POST",
  headers: {
    "Authorization": "Bearer sk-your-key",
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    model: "deepseek-v3",
    messages: [{ role: "user", content: "Hello!" }],
    stream: true,
  }),
});

const reader = response.body!.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  const text = decoder.decode(value);
  for (const line of text.split("\\n")) {
    if (line.startsWith("data: ") && line !== "data: [DONE]") {
      const json = JSON.parse(line.slice(6));
      const content = json.choices[0]?.delta?.content;
      if (content) process.stdout.write(content);
    }
  }
}
\`\`\``,
  },
  {
    id: "sdks",
    title: "SDK Examples",
    content: `__SDK_TAB_SECTION__`,
  },
  {
    id: "functions",
    title: "Function Calling",
    content: `## Function Calling

Define tools that the model can invoke. The API follows the OpenAI function calling format.

### Request with Tools

\`\`\`json
{
  "model": "deepseek-v3",
  "messages": [
    {"role": "user", "content": "What is the weather in Tokyo?"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_weather",
        "description": "Get current weather for a location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string", "description": "City name"},
            "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
          },
          "required": ["location"]
        }
      }
    }
  ],
  "tool_choice": "auto"
}
\`\`\`

### Response (tool call)

\`\`\`json
{
  "id": "chatcmpl-xxx",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": null,
      "tool_calls": [{
        "id": "call_abc123",
        "type": "function",
        "function": {
          "name": "get_weather",
          "arguments": "{\\"location\\":\\"Tokyo\\",\\"unit\\":\\"celsius\\"}"
        }
      }]
    },
    "finish_reason": "tool_calls"
  }]
}
\`\`\`

### Sending Tool Results

After receiving a tool call, execute the function and send the result back:

\`\`\`json
{
  "model": "deepseek-v3",
  "messages": [
    {"role": "user", "content": "What is the weather in Tokyo?"},
    {"role": "assistant", "content": null, "tool_calls": [{"id": "call_abc123", "type": "function", "function": {"name": "get_weather", "arguments": "{\\"location\\":\\"Tokyo\\"}"}}]},
    {"role": "tool", "tool_call_id": "call_abc123", "content": "{\\"temp\\":22,\\"condition\\":\\"sunny\\"}"}
  ]
}
\`\`\`

### Python Example

\`\`\`python
response = client.chat.completions.create(
    model="deepseek-v3",
    messages=[{"role": "user", "content": "What is the weather in Tokyo?"}],
    tools=[{
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Get current weather for a location",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {"type": "string"},
                },
                "required": ["location"],
            },
        },
    }],
)

if response.choices[0].message.tool_calls:
    tool_call = response.choices[0].message.tool_calls[0]
    print(f"Call {tool_call.function.name}({tool_call.function.arguments})")
\`\`\``,
  },
  {
    id: "errors",
    title: "Error Handling",
    content: `## Error Responses

All errors follow this format:

\`\`\`json
{
  "error": {
    "message": "description of the error",
    "type": "error_type"
  }
}
\`\`\`

### Error Types

| HTTP Code | Type | Description |
|-----------|------|-------------|
| 400 | invalid_request_error | Malformed request body or missing required fields |
| 401 | authentication_error | Invalid or missing API key |
| 402 | billing_error | Insufficient balance - top up your account |
| 404 | invalid_request_error | Model not found or endpoint does not exist |
| 413 | invalid_request_error | Request body too large (max 10MB) |
| 422 | invalid_request_error | Valid JSON but invalid parameter values |
| 429 | rate_limit_error | Rate limit exceeded - retry after backoff |
| 500 | internal_error | Unexpected server error - retry the request |
| 502 | upstream_error | Provider returned an error or is unavailable |
| 503 | service_unavailable | Service temporarily overloaded |
| 504 | timeout_error | Request timed out waiting for provider |

### Retry Strategy

For transient errors (429, 500, 502, 503, 504), implement exponential backoff:

\`\`\`python
import time
import random

def request_with_retry(func, max_retries=3):
    for attempt in range(max_retries):
        try:
            return func()
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            status = getattr(e, 'status_code', 0)
            if status not in (429, 500, 502, 503, 504):
                raise
            wait = (2 ** attempt) + random.random()
            time.sleep(wait)
\`\`\`

### Error Handling in JavaScript

\`\`\`typescript
try {
  const response = await client.chat.completions.create({
    model: "deepseek-v3",
    messages: [{ role: "user", content: "Hello!" }],
  });
} catch (error) {
  if (error instanceof OpenAI.APIError) {
    console.error("Status:", error.status);
    console.error("Message:", error.message);
    if (error.status === 429) {
      // Handle rate limiting - wait and retry
    }
  }
}
\`\`\``,
  },
  {
    id: "ratelimits",
    title: "Rate Limits",
    content: `## Rate Limits

UniLLM enforces rate limits to ensure fair usage and platform stability.

### Default Limits

| Tier | Requests per Minute | Tokens per Minute | Concurrent Requests |
|------|--------------------|--------------------|---------------------|
| Free | 20 | 40,000 | 5 |
| Standard | 200 | 400,000 | 20 |
| Pro | 1,000 | 2,000,000 | 50 |

### Rate Limit Headers

Every API response includes rate limit headers:

\`\`\`
X-RateLimit-Limit: 200
X-RateLimit-Remaining: 195
X-RateLimit-Reset: 1700000060
\`\`\`

### Handling Rate Limits

When you exceed the rate limit, you receive a \`429\` status code. The response includes a \`Retry-After\` header indicating how many seconds to wait.

\`\`\`json
{
  "error": {
    "message": "Rate limit exceeded. Please retry after 3 seconds.",
    "type": "rate_limit_error"
  }
}
\`\`\`

### Best Practices

1. **Monitor headers** - Track \`X-RateLimit-Remaining\` to proactively throttle
2. **Implement backoff** - Use exponential backoff with jitter on 429 errors
3. **Batch requests** - Combine multiple messages into a single conversation
4. **Cache responses** - Avoid repeated identical requests
5. **Use streaming** - Streaming counts as a single request regardless of output length`,
  },
  {
    id: "pricing",
    title: "Pricing",
    content: `## Pricing

Pricing is based on token usage per model. Check the dashboard for current rates.

| Model | Input (per 1M tokens) | Output (per 1M tokens) |
|-------|----------------------|----------------------|
| deepseek-v3 | varies | varies |
| claude-haiku | varies | varies |
| gemini-flash | varies | varies |

New accounts receive **$1.00 free credit**.

Balance can be topped up by contacting an administrator.`,
  },
];

export default function DocsPage() {
  const [active, setActive] = useState("quickstart");
  const [sdkLang, setSdkLang] = useState<SdkLanguage>("Python");

  const section = SECTIONS.find((s) => s.id === active);

  return (
    <div className="min-h-screen flex flex-col">
      <header
        className="border-b px-6 py-3 flex items-center gap-4"
        style={{ borderColor: "var(--border)" }}
      >
        <a href="/dashboard" className="text-lg font-bold hover:opacity-80">
          UniLLM
        </a>
        <span className="text-sm text-[var(--muted)]">API Documentation</span>
        <div className="flex-1" />
        <a
          href="/calculator"
          className="text-sm hover:opacity-80 transition-opacity"
          style={{ color: "var(--primary)" }}
        >
          Calculator
        </a>
        <a
          href="/dashboard"
          className="text-sm hover:opacity-80 transition-opacity"
          style={{ color: "var(--muted)" }}
        >
          Dashboard
        </a>
      </header>

      <div className="flex flex-1 max-w-6xl mx-auto w-full">
        {/* Sidebar */}
        <nav
          className="w-48 shrink-0 border-r p-4 space-y-1"
          style={{ borderColor: "var(--border)" }}
        >
          {SECTIONS.map((s) => (
            <button
              key={s.id}
              onClick={() => setActive(s.id)}
              className="block w-full text-left px-3 py-1.5 rounded text-sm transition-colors"
              style={{
                background: active === s.id ? "var(--primary)" : "transparent",
                color: active === s.id ? "white" : "var(--muted)",
              }}
            >
              {s.title}
            </button>
          ))}
        </nav>

        {/* Content */}
        <main className="flex-1 p-8 max-w-3xl">
          {section && section.id === "sdks" ? (
            <SdkSection selectedLang={sdkLang} onSelectLang={setSdkLang} />
          ) : (
            section && <MarkdownContent content={section.content} />
          )}
        </main>
      </div>
    </div>
  );
}

function SdkSection({
  selectedLang,
  onSelectLang,
}: {
  selectedLang: SdkLanguage;
  onSelectLang: (lang: SdkLanguage) => void;
}) {
  const example = SDK_EXAMPLES[selectedLang];

  return (
    <div>
      <h2 className="text-xl font-bold mb-4 mt-2">SDK Examples</h2>
      <p className="text-sm mb-4 leading-relaxed" style={{ color: "var(--muted)" }}>
        UniLLM is compatible with any OpenAI SDK or HTTP client. Select your
        preferred language below.
      </p>

      {/* Language Tab Selector */}
      <div
        className="flex flex-wrap gap-1 mb-4 p-1 rounded-lg"
        style={{ background: "var(--card)", border: "1px solid var(--border)" }}
      >
        {SDK_LANGUAGES.map((lang) => (
          <button
            key={lang}
            onClick={() => onSelectLang(lang)}
            className="px-3 py-1.5 rounded text-xs font-medium transition-colors"
            style={{
              background:
                selectedLang === lang ? "var(--primary)" : "transparent",
              color: selectedLang === lang ? "white" : "var(--muted)",
            }}
          >
            {lang}
          </button>
        ))}
      </div>

      {/* Code Example */}
      <pre
        className="rounded-lg p-4 text-xs overflow-x-auto mb-4 font-mono"
        style={{
          background: "#1a1a2e",
          border: "1px solid var(--border)",
        }}
      >
        <div className="text-[10px] text-[var(--muted)] mb-2 uppercase">
          {example.lang}
        </div>
        <code>{example.code}</code>
      </pre>
    </div>
  );
}

function MarkdownContent({ content }: { content: string }) {
  const lines = content.split("\n");
  const elements: React.ReactNode[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Code block
    if (line.startsWith("```")) {
      const lang = line.slice(3).trim();
      const codeLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].startsWith("```")) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // skip closing ```
      elements.push(
        <pre
          key={elements.length}
          className="rounded-lg p-4 text-xs overflow-x-auto mb-4 font-mono"
          style={{
            background: "#1a1a2e",
            border: "1px solid var(--border)",
          }}
        >
          {lang && (
            <div className="text-[10px] text-[var(--muted)] mb-2 uppercase">
              {lang}
            </div>
          )}
          <code>{codeLines.join("\n")}</code>
        </pre>
      );
      continue;
    }

    // Heading
    if (line.startsWith("## ")) {
      elements.push(
        <h2
          key={elements.length}
          className="text-xl font-bold mb-4 mt-2"
        >
          {line.slice(3)}
        </h2>
      );
      i++;
      continue;
    }
    if (line.startsWith("### ")) {
      elements.push(
        <h3
          key={elements.length}
          className="text-base font-semibold mb-3 mt-4"
        >
          {line.slice(4)}
        </h3>
      );
      i++;
      continue;
    }

    // Table
    if (line.includes("|") && lines[i + 1]?.includes("---")) {
      const headers = line.split("|").filter(Boolean).map((h) => h.trim());
      i += 2; // skip header and separator
      const rows: string[][] = [];
      while (i < lines.length && lines[i].includes("|")) {
        rows.push(lines[i].split("|").filter(Boolean).map((c) => c.trim()));
        i++;
      }
      elements.push(
        <div key={elements.length} className="overflow-x-auto mb-4">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-[var(--muted)]">
                {headers.map((h, hi) => (
                  <th key={hi} className="pb-2 pr-4">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((row, ri) => (
                <tr
                  key={ri}
                  className="border-t"
                  style={{ borderColor: "var(--border)" }}
                >
                  {row.map((cell, ci) => (
                    <td key={ci} className="py-2 pr-4 font-mono text-xs">
                      {cell}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
      continue;
    }

    // Paragraph
    if (line.trim()) {
      elements.push(
        <p key={elements.length} className="text-sm mb-3 leading-relaxed">
          {renderInline(line)}
        </p>
      );
    }
    i++;
  }

  return <>{elements}</>;
}

function renderInline(text: string): React.ReactNode[] {
  const parts: React.ReactNode[] = [];
  const regex = /`([^`]+)`|\*\*([^*]+)\*\*/g;
  let lastIndex = 0;
  let match;

  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index));
    }
    if (match[1]) {
      parts.push(
        <code
          key={parts.length}
          className="px-1 py-0.5 rounded text-xs font-mono"
          style={{ background: "var(--card)", border: "1px solid var(--border)" }}
        >
          {match[1]}
        </code>
      );
    } else if (match[2]) {
      parts.push(<strong key={parts.length}>{match[2]}</strong>);
    }
    lastIndex = match.index + match[0].length;
  }
  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }
  return parts;
}
