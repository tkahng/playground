import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { MoveSelection } from "../move";

describe("MoveSelection", () => {
  it("renders all three move options", () => {
    render(<MoveSelection handleSubmit={vi.fn()} />);
    expect(screen.getByText("Rock")).toBeInTheDocument();
    expect(screen.getByText("Paper")).toBeInTheDocument();
    expect(screen.getByText("Scissors")).toBeInTheDocument();
  });

  it("submit button disabled until a move is selected", () => {
    render(<MoveSelection handleSubmit={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Select a Move" })).toBeDisabled();
  });

  it("submit button enabled and labelled after selecting a move", async () => {
    const user = userEvent.setup();
    render(<MoveSelection handleSubmit={vi.fn()} />);

    await user.click(screen.getByText("Rock"));

    expect(
      screen.getByRole("button", { name: "Play Rock" })
    ).not.toBeDisabled();
  });

  it("calls handleSubmit with selected move on submit click", async () => {
    const handleSubmit = vi.fn();
    const user = userEvent.setup();
    render(<MoveSelection handleSubmit={handleSubmit} />);

    await user.click(screen.getByText("Paper"));
    await user.click(screen.getByRole("button", { name: "Play Paper" }));

    expect(handleSubmit).toHaveBeenCalledWith("paper");
  });

  it("submit button stays disabled when disabled=true even after selecting a move", async () => {
    const user = userEvent.setup();
    render(<MoveSelection handleSubmit={vi.fn()} disabled />);

    await user.click(screen.getByText("Scissors"));

    expect(
      screen.getByRole("button", { name: "Play Scissors" })
    ).toBeDisabled();
  });

  it("does not call handleSubmit when disabled", async () => {
    const handleSubmit = vi.fn();
    const user = userEvent.setup();
    render(<MoveSelection handleSubmit={handleSubmit} disabled />);

    await user.click(screen.getByText("Rock"));
    const btn = screen.getByRole("button", { name: "Play Rock" });
    await user.click(btn);

    expect(handleSubmit).not.toHaveBeenCalled();
  });

  it("renders children slot", () => {
    render(
      <MoveSelection handleSubmit={vi.fn()}>
        <p>custom slot content</p>
      </MoveSelection>
    );
    expect(screen.getByText("custom slot content")).toBeInTheDocument();
  });

  it("shows opponent email when opponentPlayer provided", () => {
    const opponent = {
      id: "p1",
      email: "rival@example.com",
    } as Parameters<typeof MoveSelection>[0]["opponentPlayer"];
    render(<MoveSelection handleSubmit={vi.fn()} opponentPlayer={opponent} />);
    expect(screen.getByText("rival@example.com")).toBeInTheDocument();
  });
});
