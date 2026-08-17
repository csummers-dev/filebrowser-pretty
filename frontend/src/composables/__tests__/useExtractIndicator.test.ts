import { describe, it, expect, beforeEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { StatusError } from "@/api/utils";

// Side-effecting collaborators are mocked; the real buildBatchDests /
// mapUnzipError / isArchivePasswordError (from utils) stay in play so we
// exercise the actual dest math + error branching.
const unzip = vi.fn();
const remove = vi.fn().mockResolvedValue(undefined);
vi.mock("@/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api")>();
  return {
    ...actual,
    files: {
      ...actual.files,
      unzip: (...a: unknown[]) => unzip(...a),
      remove: (...a: unknown[]) => remove(...a),
    },
  };
});

const toast = {
  info: vi.fn(() => "pid"),
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  dismiss: vi.fn(),
};
vi.mock("vue-toastification", () => ({ useToast: () => toast }));

const push = vi.fn();
vi.mock("vue-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("vue-router")>();
  return { ...actual, useRouter: () => ({ push }) };
});

const requestPassword = vi.fn();
vi.mock("@/composables/useArchivePassword", () => ({
  useArchivePassword: () => ({ requestPassword }),
}));

import { useExtractIndicator } from "@/composables/useExtractIndicator";

beforeEach(() => {
  setActivePinia(createPinia());
  vi.clearAllMocks();
  unzip.mockResolvedValue(undefined);
  remove.mockResolvedValue(undefined);
  toast.info.mockReturnValue("pid");
});

const arch = (name: string) => ({ url: `/files/${name}`, name });

describe("runExtractBatch", () => {
  it("extracts each archive into its own subfolder under the base", async () => {
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("a.zip"), arch("b.7z")], {
      base: "/files/dl/",
      overwrite: false,
      deleteOriginal: false,
      openFolder: false,
    });
    expect(unzip).toHaveBeenCalledTimes(2);
    expect(unzip).toHaveBeenNthCalledWith(
      1,
      "/files/a.zip",
      "/files/dl/a",
      false,
      undefined
    );
    expect(unzip).toHaveBeenNthCalledWith(
      2,
      "/files/b.7z",
      "/files/dl/b",
      false,
      undefined
    );
    expect(toast.success).toHaveBeenCalledWith(
      expect.stringContaining("Extracted 2 archives")
    );
  });

  it("de-duplicates colliding subfolders across archives", async () => {
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("data.zip"), arch("data.7z")], {
      base: "/x/",
      overwrite: false,
      deleteOriginal: false,
      openFolder: false,
    });
    expect(unzip).toHaveBeenNthCalledWith(
      1,
      "/files/data.zip",
      "/x/data",
      false,
      undefined
    );
    expect(unzip).toHaveBeenNthCalledWith(
      2,
      "/files/data.7z",
      "/x/data%20(2)",
      false,
      undefined
    );
  });

  it("deletes the original ONLY for archives that extracted successfully", async () => {
    // First archive fails (non-password error), second succeeds.
    unzip
      .mockRejectedValueOnce(new StatusError("boom", 500))
      .mockResolvedValueOnce(undefined);
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("bad.zip"), arch("good.zip")], {
      base: "/x/",
      overwrite: false,
      deleteOriginal: true,
      openFolder: false,
    });
    // only the successful one's source is removed
    expect(remove).toHaveBeenCalledTimes(1);
    expect(remove).toHaveBeenCalledWith("/files/good.zip");
    expect(toast.error).toHaveBeenCalledWith(
      expect.stringContaining("bad.zip")
    );
    expect(toast.success).toHaveBeenCalledWith(
      expect.stringContaining("Extracted 1 of 2")
    );
  });

  it("keeps going after a failure and reports the partial count", async () => {
    unzip.mockRejectedValue(new StatusError("nope", 500));
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("a.zip"), arch("b.zip")], {
      base: "/x/",
      overwrite: false,
      deleteOriginal: false,
      openFolder: false,
    });
    expect(unzip).toHaveBeenCalledTimes(2); // did not bail after the first
    expect(toast.success).not.toHaveBeenCalled(); // ok === 0
  });

  it("navigates to the base folder once when openFolder is on and ≥1 succeeded", async () => {
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("a.zip"), arch("b.zip")], {
      base: "/x",
      overwrite: false,
      deleteOriginal: false,
      openFolder: true,
    });
    expect(push).toHaveBeenCalledTimes(1);
    expect(push).toHaveBeenCalledWith({ path: "/x/" });
  });

  it("does not navigate when nothing extracted", async () => {
    unzip.mockRejectedValue(new StatusError("nope", 500));
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("a.zip")], {
      base: "/x/",
      overwrite: false,
      deleteOriginal: false,
      openFolder: true,
    });
    expect(push).not.toHaveBeenCalled();
  });

  it("passes the overwrite flag through to the API", async () => {
    const { runExtractBatch } = useExtractIndicator();
    await runExtractBatch([arch("a.zip")], {
      base: "/x/",
      overwrite: true,
      deleteOriginal: false,
      openFolder: false,
    });
    expect(unzip).toHaveBeenCalledWith("/files/a.zip", "/x/a", true, undefined);
  });
});
