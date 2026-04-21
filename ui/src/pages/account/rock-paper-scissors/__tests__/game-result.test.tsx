import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { GameResult } from "../game-result";

describe("GameResult", () => {
  const baseProps = {
    result: "win" as const,
    opponent: "alice@example.com",
    playerMove: "rock" as const,
    opponentMove: "scissors" as const,
  };

  describe("result header", () => {
    it("shows Victory on win", () => {
      render(<GameResult {...baseProps} result="win" />);
      expect(screen.getByText("Victory!")).toBeInTheDocument();
    });

    it("shows Defeat on loss", () => {
      render(<GameResult {...baseProps} result="lose" />);
      expect(screen.getByText("Defeat")).toBeInTheDocument();
    });

    it("shows Tie on tie", () => {
      render(<GameResult {...baseProps} result="tie" />);
      expect(screen.getByText("It's a Tie!")).toBeInTheDocument();
    });

    it("shows opponent name", () => {
      render(<GameResult {...baseProps} opponent="bob@example.com" />);
      expect(screen.getByText("bob@example.com")).toBeInTheDocument();
    });
  });

  describe("move display", () => {
    it("shows player and opponent moves", () => {
      render(
        <GameResult {...baseProps} playerMove="rock" opponentMove="scissors" />
      );
      // jsdom does not apply CSS capitalize — text nodes are lowercase
      expect(screen.getByText("rock")).toBeInTheDocument();
      expect(screen.getByText("scissors")).toBeInTheDocument();
    });
  });

  describe("bet outcome card", () => {
    it("hidden when no betAmount", () => {
      render(<GameResult {...baseProps} />);
      expect(screen.queryByText(/pts/)).not.toBeInTheDocument();
    });

    it("hidden when betAmount present but betResult missing", () => {
      render(<GameResult {...baseProps} betAmount={50} />);
      expect(screen.queryByText(/pts/)).not.toBeInTheDocument();
    });

    it("shows +pts and 'Bet won' on win", () => {
      render(
        <GameResult {...baseProps} betAmount={100} betResult="win" />
      );
      expect(screen.getByText("+100 pts")).toBeInTheDocument();
      expect(screen.getByText("Bet won")).toBeInTheDocument();
    });

    it("shows −pts and 'Bet lost' on loss", () => {
      render(
        <GameResult {...baseProps} betAmount={50} betResult="lose" />
      );
      expect(screen.getByText("−50 pts")).toBeInTheDocument();
      expect(screen.getByText("Bet lost")).toBeInTheDocument();
    });

    it("shows pts and 'Bet refunded' on tie", () => {
      render(
        <GameResult {...baseProps} betAmount={75} betResult="tie" />
      );
      expect(screen.getByText("75 pts")).toBeInTheDocument();
      expect(screen.getByText("Bet refunded")).toBeInTheDocument();
    });
  });
});
