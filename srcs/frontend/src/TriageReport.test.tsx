import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { TriageReport } from "./TriageReport";

describe("TriageReport Component", () => {
  it("renders correctly with given glassmorphism CSS classes and metrics", () => {
    const { container } = render(<TriageReport />);

    // Verify root classes
    expect(container.querySelector(".triage-report")).toBeInTheDocument();
    expect(container.querySelector(".triage-report-glass")).toBeInTheDocument();

    // Verify text
    expect(screen.getByText("Swarm Hygiene Report")).toBeInTheDocument();
    expect(screen.getByText("The backlog is clean, prioritized, and correctly labeled.")).toBeInTheDocument();

    // Verify metrics exist
    expect(screen.getByText("Stale Missions Pruned")).toBeInTheDocument();
    expect(screen.getByText("Signal Noise Filtered")).toBeInTheDocument();
    expect(screen.getByText("Test Coverage")).toBeInTheDocument();

    // Verify values exist
    expect(screen.getByText("100%")).toBeInTheDocument();
    expect(screen.getByText("Yes")).toBeInTheDocument();
    expect(screen.getByText(">95%")).toBeInTheDocument();
  });
});
