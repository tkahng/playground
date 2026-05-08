import { describe, expect, it } from "vitest";
import { features } from "../features-list";

describe("features list", () => {
  describe("discovery order", () => {
    it("has 7 features", () => {
      expect(features).toHaveLength(7);
    });

    it("Say Hello is first", () => {
      expect(features[0]!.fragment).toBe("say-hello");
    });

    it("Authentication is second", () => {
      expect(features[1]!.fragment).toBe("authentication");
    });

    it("Teams is third", () => {
      expect(features[2]!.fragment).toBe("teams");
    });

    it("Projects & Tasks is fourth", () => {
      expect(features[3]!.fragment).toBe("projects-and-tasks");
    });

    it("Plans/Billing is fifth", () => {
      expect(features[4]!.fragment).toBe("payment-integration");
    });

    it("Rock Paper Scissors is sixth", () => {
      expect(features[5]!.fragment).toBe("rps");
    });

    it("Admin is last", () => {
      expect(features[6]!.fragment).toBe("admin");
    });
  });

  describe("fragment ids are unique", () => {
    it("no duplicate fragment values", () => {
      const fragments = features.map((f) => f.fragment);
      expect(new Set(fragments).size).toBe(features.length);
    });
  });

  describe("badges", () => {
    it("every feature has a badge", () => {
      features.forEach((f) => {
        expect(f.badge, `${f.title} is missing a badge`).toBeTruthy();
      });
    });

    it("Say Hello badge signals no sign-up needed", () => {
      expect(features[0]!.badge?.toLowerCase()).toContain("no sign-up");
    });

    it("Admin badge signals role requirement", () => {
      expect(features[6]!.badge?.toLowerCase()).toContain("admin");
    });
  });

  describe("action links", () => {
    it("Say Hello detailLink goes to /say-hello", () => {
      expect(features[0]!.detailLink).toBe("/say-hello");
    });

    it("Authentication detailLink goes to /signup", () => {
      expect(features[1]!.detailLink).toBe("/signup");
    });

    it("Plans detailLink goes to /pricing", () => {
      expect(features[4]!.detailLink).toBe("/pricing");
    });

    it("RPS detailLink goes to /rock-paper-scissors", () => {
      expect(features[5]!.detailLink).toBe("/rock-paper-scissors");
    });
  });
});
