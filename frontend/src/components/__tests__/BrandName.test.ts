import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import BrandName from "@/components/BrandName.vue";

// BrandName is the app wordmark. When the name matches "vitrine"
// (case-insensitive, surrounding whitespace ignored) every letter renders in
// its own `.brand-name__ltr` span with a cycling accent class (--0..--5);
// anything else renders as plain text. These lock that behavior — it's the
// brand identity and the only place the per-letter accent logic lives.

const accentLetters = (wrapper: ReturnType<typeof mount>) =>
  wrapper.findAll(".brand-name__ltr");

describe("BrandName", () => {
  it("accents every letter of the wordmark", () => {
    const wrapper = mount(BrandName, { props: { name: "vitrine" } });
    const letters = accentLetters(wrapper);
    expect(letters).toHaveLength("vitrine".length);
    expect(letters.map((l) => l.text()).join("")).toBe("vitrine");
  });

  it("cycles the six accent classes across the letters", () => {
    const wrapper = mount(BrandName, { props: { name: "vitrine" } });
    const classes = accentLetters(wrapper).map((l) =>
      // the modulo-6 suffix class, e.g. "brand-name__ltr--3"
      l.classes().find((c) => c.startsWith("brand-name__ltr--"))
    );
    // 7 letters → indices 0,1,2,3,4,5,0 (cycles at the 7th)
    expect(classes).toEqual([
      "brand-name__ltr--0",
      "brand-name__ltr--1",
      "brand-name__ltr--2",
      "brand-name__ltr--3",
      "brand-name__ltr--4",
      "brand-name__ltr--5",
      "brand-name__ltr--0",
    ]);
  });

  it("matches case-insensitively but preserves the user's casing", () => {
    const wrapper = mount(BrandName, { props: { name: "Vitrine" } });
    const letters = accentLetters(wrapper);
    expect(letters).toHaveLength("Vitrine".length);
    expect(letters.map((l) => l.text()).join("")).toBe("Vitrine");
  });

  it("ignores surrounding whitespace when matching", () => {
    const wrapper = mount(BrandName, { props: { name: "  vitrine  " } });
    expect(accentLetters(wrapper)).toHaveLength("vitrine".length);
    // the full string (incl. spaces) is still present
    expect(wrapper.text()).toBe("vitrine");
  });

  it("renders a custom instance name as plain text (no accent)", () => {
    const wrapper = mount(BrandName, { props: { name: "My Files" } });
    expect(accentLetters(wrapper)).toHaveLength(0);
    expect(wrapper.text()).toBe("My Files");
  });

  it("does not accent a name that merely contains the wordmark", () => {
    const wrapper = mount(BrandName, { props: { name: "My vitrine" } });
    expect(accentLetters(wrapper)).toHaveLength(0);
    expect(wrapper.text()).toBe("My vitrine");
  });
});
