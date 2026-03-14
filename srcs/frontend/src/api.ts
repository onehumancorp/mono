import type { CostSummary, MeetingRoom, Organization } from "./types";

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`Request failed for ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

export function fetchOrganization(): Promise<Organization> {
  return getJSON<Organization>("/api/org");
}

export function fetchMeetings(): Promise<MeetingRoom[]> {
  return getJSON<MeetingRoom[]>("/api/meetings").then((meetings) =>
    meetings.map((meeting) => ({
      ...meeting,
      transcript: meeting.transcript ?? [],
    }))
  );
}

export async function fetchCosts(): Promise<CostSummary> {
  const response = await getJSON<Record<string, unknown>>("/api/costs");
  const agents = Array.isArray(response.agents) ? response.agents : [];

  return {
    organizationID: String(response.organizationID ?? response.organizationId ?? ""),
    totalTokens: Number(response.totalTokens ?? 0),
    totalCostUSD: Number(response.totalCostUSD ?? response.totalCostUsd ?? 0),
    agents: agents.map((agent) => {
      const value = agent as Record<string, unknown>;
      return {
        agentID: String(value.agentID ?? value.agentId ?? ""),
        model: String(value.model ?? ""),
        tokenUsed: Number(value.tokenUsed ?? 0),
        costUSD: Number(value.costUSD ?? value.costUsd ?? 0),
      };
    }),
  };
}

export async function sendMessage(form: {
  fromAgent: string;
  toAgent: string;
  meetingId: string;
  messageType: string;
  content: string;
}): Promise<void> {
  const params = new URLSearchParams(form);
  const response = await fetch("/api/messages", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: params.toString(),
    redirect: "follow",
  });

  if (!response.ok) {
    throw new Error(`Failed to send message: ${response.status}`);
  }
}