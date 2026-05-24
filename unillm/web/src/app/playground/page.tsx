"use client";

import { useEffect, useState, useRef } from "react";
import { useRouter } from "next/navigation";
import { isLoggedIn } from "@/lib/api";

interface Message {
  role: "system" | "user" | "assistant";
  content: string;
}

export default function PlaygroundPage() {
  const router = useRouter();
  const [apiKey, setApiKey] = useState("");
  const [model, setModel] = useState("deepseek-chat");
  const [models, setModels] = useState<string[]>([]);
  const [systemPrompt, setSystemPrompt] = useState("You are a helpful assistant.");
  const [input, setInput] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [streaming, setStreaming] = useState(false);
  const [streamText, setStreamText] = useState("");
  const outputRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isLoggedIn()) {
      router.push("/login");
      return;
    }
    const saved = localStorage.getItem("playground_key");
    if (saved) {
      setApiKey(saved);
      loadModels(saved);
    }
  }, [router]);

  async function loadModels(key: string) {
    try {
      const res = await fetch("/v1/models", {
        headers: { Authorization: `Bearer ${key}` },
      });
      const data = await res.json();
      const ids = (data.data || []).map((m: { id: string }) => m.id);
      setModels(ids);
      if (ids.length > 0 && !ids.includes(model)) {
        setModel(ids[0]);
      }
    } catch {
      // ignore
    }
  }

  function handleKeyChange(key: string) {
    setApiKey(key);
    localStorage.setItem("playground_key", key);
    if (key.startsWith("sk-")) loadModels(key);
  }

  async function handleSend() {
    if (!input.trim() || !apiKey || streaming) return;

    const userMsg: Message = { role: "user", content: input.trim() };
    const allMessages: Message[] = [
      ...(systemPrompt ? [{ role: "system" as const, content: systemPrompt }] : []),
      ...messages,
      userMsg,
    ];

    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setStreaming(true);
    setStreamText("");

    try {
      const res = await fetch("/v1/chat/completions", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify({
          model,
          messages: allMessages,
          stream: true,
        }),
      });

      if (!res.ok) {
        const err = await res.text();
        setMessages((prev) => [
          ...prev,
          { role: "assistant", content: `Error: ${err}` },
        ]);
        setStreaming(false);
        return;
      }

      const reader = res.body?.getReader();
      const decoder = new TextDecoder();
      let fullText = "";

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          const lines = chunk.split("\n");

          for (const line of lines) {
            if (!line.startsWith("data: ")) continue;
            const data = line.slice(6).trim();
            if (data === "[DONE]") continue;

            try {
              const parsed = JSON.parse(data);
              const delta = parsed.choices?.[0]?.delta?.content;
              if (delta) {
                fullText += delta;
                setStreamText(fullText);
              }
            } catch {
              // skip unparseable chunks
            }
          }
        }
      }

      setMessages((prev) => [
        ...prev,
        { role: "assistant", content: fullText || "(empty response)" },
      ]);
      setStreamText("");
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        {
          role: "assistant",
          content: `Error: ${err instanceof Error ? err.message : "Request failed"}`,
        },
      ]);
    } finally {
      setStreaming(false);
    }
  }

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [messages, streamText]);

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header
        className="border-b px-6 py-3 flex items-center justify-between"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="flex items-center gap-4">
          <a href="/dashboard" className="text-lg font-bold hover:opacity-80">
            UniLLM
          </a>
          <a
            href="/models"
            className="text-sm text-[var(--muted)] hover:text-white transition-colors"
          >
            Models
          </a>
          <span className="text-sm text-[var(--muted)]">Playground</span>
        </div>
      </header>

      <div className="flex-1 flex flex-col max-w-4xl mx-auto w-full p-6 gap-4">
        {/* Config bar */}
        <div className="flex flex-wrap gap-3 items-end">
          <div className="flex-1 min-w-[200px]">
            <label className="block text-xs text-[var(--muted)] mb-1">
              API Key
            </label>
            <input
              type="password"
              value={apiKey}
              onChange={(e) => handleKeyChange(e.target.value)}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: "var(--card)",
                border: "1px solid var(--border)",
              }}
              placeholder="Paste your API key (sk-...)"
            />
          </div>
          <div className="min-w-[200px]">
            <label className="block text-xs text-[var(--muted)] mb-1">
              Model
            </label>
            <select
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: "var(--card)",
                border: "1px solid var(--border)",
                color: "var(--foreground)",
              }}
            >
              {models.length > 0 ? (
                models.map((m) => (
                  <option key={m} value={m}>
                    {m}
                  </option>
                ))
              ) : (
                <option value={model}>{model}</option>
              )}
            </select>
          </div>
          <button
            onClick={() => {
              setMessages([]);
              setStreamText("");
            }}
            className="px-3 py-2 rounded-lg text-sm"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
            }}
          >
            Clear
          </button>
        </div>

        {/* System prompt */}
        <div>
          <label className="block text-xs text-[var(--muted)] mb-1">
            System Prompt
          </label>
          <input
            value={systemPrompt}
            onChange={(e) => setSystemPrompt(e.target.value)}
            className="w-full px-3 py-2 rounded-lg text-sm outline-none"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
            }}
            placeholder="System prompt..."
          />
        </div>

        {/* Messages */}
        <div
          ref={outputRef}
          className="flex-1 min-h-[300px] max-h-[500px] overflow-y-auto rounded-xl p-4 space-y-4"
          style={{
            background: "var(--card)",
            border: "1px solid var(--border)",
          }}
        >
          {messages.length === 0 && !streamText && (
            <p className="text-sm text-[var(--muted)] text-center mt-8">
              Send a message to get started
            </p>
          )}
          {messages.map((m, i) => (
            <div key={i}>
              <div
                className="text-xs font-medium mb-1"
                style={{
                  color:
                    m.role === "user" ? "var(--primary)" : "var(--success)",
                }}
              >
                {m.role === "user" ? "You" : "Assistant"}
              </div>
              <div className="text-sm whitespace-pre-wrap">{m.content}</div>
            </div>
          ))}
          {streaming && streamText && (
            <div>
              <div
                className="text-xs font-medium mb-1"
                style={{ color: "var(--success)" }}
              >
                Assistant
              </div>
              <div className="text-sm whitespace-pre-wrap">{streamText}</div>
            </div>
          )}
          {streaming && !streamText && (
            <div className="text-sm text-[var(--muted)]">Thinking...</div>
          )}
        </div>

        {/* Input */}
        <div className="flex gap-2">
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            placeholder={
              apiKey
                ? "Type a message..."
                : "Select an API key first"
            }
            disabled={!apiKey || streaming}
            className="flex-1 px-4 py-3 rounded-lg text-sm outline-none"
            style={{
              background: "var(--background)",
              border: "1px solid var(--border)",
            }}
          />
          <button
            onClick={handleSend}
            disabled={!apiKey || !input.trim() || streaming}
            className="px-6 py-3 rounded-lg text-sm font-medium text-white"
            style={{
              background:
                !apiKey || !input.trim() || streaming
                  ? "var(--muted)"
                  : "var(--primary)",
            }}
          >
            Send
          </button>
        </div>
      </div>
    </div>
  );
}
