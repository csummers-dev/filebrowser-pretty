<template>
  <main class="login">
    <!-- Decorative mesh layer — sits behind the card, never receives clicks -->
    <div class="login__mesh" aria-hidden="true"></div>

    <form class="login-card" @submit="submit" novalidate>
      <!-- Header: workspace logo + name chip -->
      <header class="login-card__header">
        <div class="login-card__logo">
          <img
            v-if="logoURL"
            :src="logoURL"
            :alt="brand"
            class="login-card__logo-img"
          />
          <Icon
            v-else
            name="folder"
            :size="16"
            :stroke-width="1.8"
            class="login-card__logo-fallback"
          />
        </div>
        <div class="login-card__brand">
          <BrandName :name="brand" />
        </div>
      </header>

      <!-- Title. The wordmark is fixed (never the stored Branding.Name), so
           the brand is always present — no empty-name fallback needed. -->
      <h1 class="login-card__title">
        <template v-if="createMode">
          Create your <BrandName :name="brand" /> account
        </template>
        <template v-else> Sign in to <BrandName :name="brand" /> </template>
      </h1>
      <p class="login-card__subtitle">
        {{
          createMode
            ? "Pick a username and password to get started."
            : "Welcome back. Enter your credentials to continue."
        }}
      </p>

      <!-- Logout reason banner (e.g. session expired) -->
      <div v-if="reason" class="login-banner">
        <Icon name="info" :size="13" />
        <span>{{ t(`login.logout_reasons.${reason}`) }}</span>
      </div>

      <!-- Form fields -->
      <div class="login-field">
        <label for="login-username" class="login-field__label">
          {{ t("login.username") }}
        </label>
        <input
          id="login-username"
          ref="usernameInput"
          v-model="username"
          name="username"
          type="text"
          autocomplete="username"
          autocapitalize="off"
          class="login-input"
          :class="{ 'is-error': !!error }"
          :disabled="submitting"
          @input="clearError"
        />
      </div>

      <div class="login-field">
        <label for="login-password" class="login-field__label">
          {{ t("login.password") }}
        </label>
        <div class="login-input-wrap">
          <input
            id="login-password"
            v-model="password"
            name="password"
            :type="showPassword ? 'text' : 'password'"
            :autocomplete="createMode ? 'new-password' : 'current-password'"
            class="login-input login-input--has-toggle"
            :class="{ 'is-error': !!error || passwordMismatch }"
            :disabled="submitting"
            @input="clearError"
          />
          <button
            type="button"
            class="login-password-toggle"
            :aria-label="showPassword ? 'Hide password' : 'Show password'"
            :aria-pressed="showPassword"
            :title="showPassword ? 'Hide password' : 'Show password'"
            :disabled="submitting"
            @click="showPassword = !showPassword"
          >
            <Icon :name="showPassword ? 'eye-off' : 'eye'" :size="15" />
          </button>
        </div>
      </div>

      <div v-if="createMode" class="login-field">
        <label for="login-password-confirm" class="login-field__label">
          {{ t("login.passwordConfirm") }}
        </label>
        <div class="login-input-wrap">
          <input
            id="login-password-confirm"
            v-model="passwordConfirm"
            name="password-confirm"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="new-password"
            class="login-input login-input--has-toggle"
            :class="passwordConfirmClass"
            :disabled="submitting"
            @input="clearError"
          />
          <button
            type="button"
            class="login-password-toggle"
            :aria-label="showPassword ? 'Hide password' : 'Show password'"
            :aria-pressed="showPassword"
            :title="showPassword ? 'Hide password' : 'Show password'"
            :disabled="submitting"
            @click="showPassword = !showPassword"
          >
            <Icon :name="showPassword ? 'eye-off' : 'eye'" :size="15" />
          </button>
        </div>
        <p v-if="passwordMismatch" class="login-helper login-helper--error">
          Passwords don't match.
        </p>
      </div>

      <!-- reCAPTCHA — preserved as-is, wrapped for layout consistency -->
      <div v-if="recaptcha" class="login-field">
        <div id="recaptcha"></div>
      </div>

      <!-- Inline error chip (appears immediately above submit) -->
      <Transition name="login-error">
        <div v-if="error" class="login-error" role="alert">
          <Icon name="triangle-alert" :size="13" />
          <span>{{ error }}</span>
        </div>
      </Transition>

      <!-- Submit button — loading state during the auth.login() await -->
      <button
        type="submit"
        class="login-submit"
        :disabled="submitting || (createMode && passwordMismatch)"
      >
        <Icon
          v-if="submitting"
          name="loader-circle"
          :size="14"
          class="login-spin"
        />
        <span>{{ submitLabel }}</span>
      </button>

      <!-- Mode toggle (signup only) -->
      <p v-if="signup" class="login-toggle">
        <template v-if="createMode">
          <span>{{ t("login.loginInstead") }}</span>
          <button type="button" class="login-toggle__link" @click="toggleMode">
            Sign in
          </button>
        </template>
        <template v-else>
          <span>Click</span>
          <button type="button" class="login-toggle__link" @click="toggleMode">
            here
          </button>
          <span>to create an account</span>
        </template>
      </p>

      <!-- App version footer — links to the project repo. -->
      <footer v-if="version" class="login-card__footer">
        <a
          :href="repoUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="login-card__version"
          >v{{ version }}</a
        >
      </footer>
    </form>
  </main>
</template>

<script setup lang="ts">
import { StatusError } from "@/api/utils";
import * as auth from "@/utils/auth";
import { settings as settingsApi } from "@/api";
import { useAuthStore } from "@/stores/auth";
import { lastFilesPathKey } from "@/router";
import {
  brand,
  recaptcha,
  recaptchaKey,
  signup,
  version,
  repoUrl,
} from "@/utils/constants";
import { useBrandLogo } from "@/composables/useBrandLogo";
import { computed, inject, nextTick, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import Icon from "@/components/Icon.vue";
import BrandName from "@/components/BrandName.vue";

const { t } = useI18n({});
const { logoURL } = useBrandLogo();

// ── Refs ─────────────────────────────────────────────────────────────
const createMode = ref<boolean>(false);
const error = ref<string>("");
const username = ref<string>("");
const password = ref<string>("");
const passwordConfirm = ref<string>("");
const submitting = ref<boolean>(false);
const usernameInput = ref<HTMLInputElement | null>(null);

// Shared visibility toggle for both password and confirm fields. One
// click reveals both — common pattern, and matches the upstream PR's
// intent. Resets to hidden on mount (default false).
const showPassword = ref<boolean>(false);

const route = useRoute();
const router = useRouter();

const $showError = inject<IToastError>("$showError")!;

const reason = route.query["logout-reason"] ?? null;

// ── Computed ─────────────────────────────────────────────────────────
// Title composition lives in the template now so the easter-egg
// <BrandName> component can render inline accent spans without losing
// the surrounding "Sign in to …" / "Create your … account" phrasing.

const submitLabel = computed(() => {
  if (submitting.value) {
    return createMode.value ? "Creating account…" : "Signing in…";
  }
  return createMode.value ? t("login.signup") : t("login.submit");
});

const passwordMismatch = computed(
  () =>
    createMode.value &&
    password.value !== "" &&
    passwordConfirm.value !== "" &&
    password.value !== passwordConfirm.value
);

const passwordConfirmClass = computed(() => {
  if (passwordConfirm.value === "") return "";
  if (passwordMismatch.value) return "is-error";
  if (password.value === passwordConfirm.value) return "is-ok";
  return "";
});

// ── Handlers ─────────────────────────────────────────────────────────
const toggleMode = () => {
  createMode.value = !createMode.value;
  clearError();
};

const clearError = () => {
  if (error.value) error.value = "";
};

const submit = async (event: Event) => {
  event.preventDefault();
  event.stopPropagation();
  if (submitting.value) return;

  const redirect = (route.query.redirect || "/files/") as string;

  let captcha = "";
  if (recaptcha) {
    captcha = window.grecaptcha.getResponse();
    if (captcha === "") {
      error.value = t("login.wrongCredentials");
      return;
    }
  }

  if (createMode.value) {
    if (password.value !== passwordConfirm.value) {
      error.value = t("login.passwordsDontMatch");
      return;
    }
  }

  submitting.value = true;
  try {
    if (createMode.value) {
      await auth.signup(username.value, password.value);
    }
    await auth.login(username.value, password.value, captcha);

    // Honor an explicit ?redirect= first. Otherwise consult the global
    // "Re-open last visited page on login" setting and (if on) restore
    // the user's last /files path.
    let destination = redirect;
    if (!route.query.redirect) {
      try {
        const cfg = await settingsApi.get();
        if (cfg.rememberLastPage) {
          const authStore = useAuthStore();
          const key = lastFilesPathKey(authStore.user?.username);
          const last = localStorage.getItem(key);
          if (last && last.startsWith("/files")) destination = last;
        }
      } catch {
        // Network blip — fall back to the default redirect
      }
    }
    router.push({ path: destination });
  } catch (e: any) {
    if (e instanceof StatusError) {
      if (e.status === 409) {
        error.value = t("login.usernameTaken");
      } else if (e.status === 403) {
        error.value = t("login.wrongCredentials");
      } else if (e.status === 400) {
        const match = e.message.match(/minimum length is (\d+)/);
        if (match) {
          error.value = t("login.passwordTooShort", { min: match[1] });
        } else {
          error.value = e.message;
        }
      } else {
        $showError(e);
      }
    } else if (e instanceof Error) {
      $showError(e);
    }
  } finally {
    submitting.value = false;
  }
};

// ── Lifecycle ────────────────────────────────────────────────────────
onMounted(async () => {
  await nextTick();
  usernameInput.value?.focus();

  if (recaptcha) {
    window.grecaptcha.ready(function () {
      window.grecaptcha.render("recaptcha", { sitekey: recaptchaKey });
    });
  }
});
</script>

<style scoped>
.login {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  background: var(--color-canvas, #fafaf9);
  position: relative;
  overflow: hidden;
  font-family: var(--font-sans, system-ui);
  color: var(--color-ink-1, #18181b);
}

/* Decorative mesh background — one soft radial blob per accent preset, so the
   page reads as a gradient of all six accent colors (lilac, blue, teal, green,
   amber, rose) arranged around the card. Hard-coded hues (not the accent CSS
   var) because login is pre-auth, where the chosen accent isn't applied yet.
   IMPORTANT: no `filter: blur()`. A full-viewport blur filter is very
   expensive to paint on every reflow/repaint (which input events trigger),
   so we rely on the radial gradients' own soft edges + tuned alpha + a
   wider spread to stay ambient without the GPU cost. */
.login__mesh {
  position: absolute;
  inset: 0;
  /* Calm Minimal: a single faint accent wash (two soft blobs) replacing the
     six-color "all accents" mesh. Uses the default accent vars (violet) since
     the picker hasn't loaded pre-auth. */
  background:
    radial-gradient(
      ellipse 64% 60% at 16% 18%,
      rgb(var(--accent-rgb) / 0.14) 0%,
      transparent 62%
    ),
    radial-gradient(
      ellipse 60% 56% at 88% 84%,
      rgb(var(--accent-rgb) / 0.1) 0%,
      transparent 62%
    );
  pointer-events: none;
  z-index: 0;
}

/* Dark mode: stronger alpha so the accent wash stays visible against the dark
   canvas. */
html.dark .login__mesh {
  background:
    radial-gradient(
      ellipse 64% 60% at 16% 18%,
      rgb(var(--accent-rgb) / 0.24) 0%,
      transparent 64%
    ),
    radial-gradient(
      ellipse 60% 56% at 88% 84%,
      rgb(var(--accent-rgb) / 0.16) 0%,
      transparent 64%
    );
}

/* ── Card ───────────────────────────────────────────────────────────── */
.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
  background: var(--color-surface, #fff);
  border: 1px solid var(--color-line, #ececec);
  border-radius: 14px;
  padding: 24px 24px 20px;
  box-shadow:
    0 32px 64px -16px rgba(0, 0, 0, 0.15),
    0 0 0 1px rgba(0, 0, 0, 0.03);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.login-card__header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.login-card__logo {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}

/* Fallback folder glyph (no logo) keeps the branded accent tile; the real
   logo.png renders clean on transparent, matching the sidebar brand mark. */
.login-card__logo:has(.login-card__logo-fallback) {
  background: linear-gradient(
    135deg,
    var(--color-accent, #6e72d9) 0%,
    var(--color-accent-strong, #575cc7) 100%
  );
  box-shadow: 0 1px 2px rgba(110, 114, 217, 0.4);
}

.login-card__logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.login-card__logo-fallback {
  color: white;
}

.login-card__brand {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--color-ink-2, #52525b);
  letter-spacing: -0.005em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.login-card__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-ink-1, #18181b);
  margin: 0;
  letter-spacing: -0.01em;
  line-height: 1.2;
}

.login-card__subtitle {
  margin: -8px 0 4px;
  font-size: 13px;
  line-height: 1.45;
  color: var(--color-ink-2, #52525b);
}

/* ── Logout-reason banner ───────────────────────────────────────────── */
.login-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  background: var(--color-accent-soft, rgba(110, 114, 217, 0.1));
  color: var(--color-accent, #6e72d9);
  font-size: 12.5px;
  font-weight: 500;
  line-height: 1.4;
}

/* ── Form fields ────────────────────────────────────────────────────── */
.login-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.login-field__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-ink-1, #18181b);
}

.login-input {
  width: 100%;
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--color-line, #ececec);
  border-radius: 8px;
  background: var(--color-surface, #fff);
  font: inherit;
  font-size: 14px;
  color: var(--color-ink-1, #18181b);
  outline: none;
  transition:
    border-color 0.1s ease,
    box-shadow 0.1s ease,
    background-color 0.1s ease;
}

.login-input:focus {
  border-color: var(--color-accent, #6e72d9);
  box-shadow: 0 0 0 3px var(--color-accent-ring, rgba(110, 114, 217, 0.3));
}

.login-input:disabled {
  background: var(--color-canvas, #fafaf9);
  color: var(--color-ink-3, #a1a1aa);
  cursor: not-allowed;
}

.login-input.is-error {
  border-color: var(--status-danger-fill);
  box-shadow: 0 0 0 3px var(--status-danger-soft);
}

.login-input.is-error:focus {
  border-color: var(--status-danger-fill);
  box-shadow: 0 0 0 3px rgba(176, 96, 96, 0.25);
}

.login-input.is-ok {
  border-color: var(--status-success);
}

/* Wrap for inputs that host an in-field action (currently the password
   visibility toggle). Establishes a positioning context and keeps the
   input filling the full row. */
.login-input-wrap {
  position: relative;
  display: block;
}

/* Extra right padding so typed text never slides under the toggle
   button. 40 px ≈ 32 px button + 4 px gap on each side. */
.login-input--has-toggle {
  padding-right: 40px;
}

html[dir="rtl"] .login-input--has-toggle {
  padding-right: 12px;
  padding-left: 40px;
}

/* The eye / eye-off button itself. 32x32 tap target satisfies the
   Stage 11f touch sweep; sits flush to the right edge of the field. */
.login-password-toggle {
  position: absolute;
  top: 50%;
  right: 4px;
  transform: translateY(-50%);
  width: 28px;
  height: 28px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--color-ink-3, #a1a1aa);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition:
    background-color var(--dur-base) ease,
    color var(--dur-base) ease,
    box-shadow 0.1s ease;
}

html[dir="rtl"] .login-password-toggle {
  right: auto;
  left: 4px;
}

.login-password-toggle:hover {
  background: var(--color-elevated, #f4f4f5);
  color: var(--color-ink-1, #18181b);
}

.login-password-toggle:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 1px;
  color: var(--color-ink-1, #18181b);
}

.login-password-toggle:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.login-helper {
  margin: 4px 0 0;
  font-size: 11.5px;
}

.login-helper--error {
  color: var(--status-danger);
}

/* ── Inline error chip (above submit) ───────────────────────────────── */
.login-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border-radius: 8px;
  background: var(--status-danger-soft);
  border: 1px solid var(--status-danger-ring);
  color: var(--status-danger);
  font-size: 12.5px;
  font-weight: 500;
  line-height: 1.4;
}

/* ── Submit ─────────────────────────────────────────────────────────── */
.login-submit {
  height: 36px;
  border-radius: 8px;
  border: 1px solid var(--color-accent, #6e72d9);
  background: var(--accent-gradient);
  color: white;
  font-family: inherit;
  font-size: 13.5px;
  font-weight: 600;
  letter-spacing: -0.005em;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 2px;
  transition:
    background-color 0.1s ease,
    border-color 0.1s ease,
    box-shadow 0.1s ease;
}

.login-submit:hover:not(:disabled) {
  background: var(--accent-gradient-strong);
  border-color: var(--color-accent-strong, #575cc7);
  box-shadow: 0 4px 12px -4px rgba(110, 114, 217, 0.5);
}

.login-submit:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 2px;
}

.login-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-spin {
  animation: login-spin 0.9s linear infinite;
}

@keyframes login-spin {
  to {
    transform: rotate(360deg);
  }
}

/* ── Mode toggle ────────────────────────────────────────────────────── */
.login-toggle {
  margin: 4px 0 0;
  font-size: 12.5px;
  color: var(--color-ink-2, #52525b);
  text-align: center;
}

.login-toggle__link {
  background: none;
  border: 0;
  padding: 0;
  font: inherit;
  font-weight: 600;
  color: var(--color-accent, #6e72d9);
  cursor: pointer;
  /* Symmetric so the link reads with a space on BOTH sides whether it sits
     at the end ("…account? Sign in") or mid-sentence ("Click here to …"). */
  margin: 0 4px;
}

/* Version footer link → project repo. Subtle by default, accent on hover. */
.login-card__version {
  color: inherit;
  text-decoration: none;
  transition: color var(--dur-base) ease;
}
.login-card__version:hover {
  color: var(--color-accent, #6e72d9);
  text-decoration: underline;
}

.login-toggle__link:hover {
  text-decoration: underline;
}

.login-toggle__link:focus-visible {
  outline: 2px solid var(--color-accent-ring, rgba(110, 114, 217, 0.3));
  outline-offset: 2px;
  border-radius: 4px;
}

/* ── Footer ─────────────────────────────────────────────────────────── */
.login-card__footer {
  margin-top: 4px;
  text-align: center;
  font-family: var(--font-mono, monospace);
  font-size: 10.5px;
  color: var(--color-ink-3, #a1a1aa);
  letter-spacing: 0.02em;
}

/* ── Transitions ────────────────────────────────────────────────────── */
.login-error-enter-active,
.login-error-leave-active {
  transition:
    opacity var(--dur-base) ease,
    transform 0.16s cubic-bezier(0.4, 0, 0.2, 1);
}
.login-error-enter-from,
.login-error-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ── Mobile tweaks ──────────────────────────────────────────────────── */
@media (max-width: 480px) {
  .login {
    padding: 20px 12px;
    align-items: flex-start;
    padding-top: 12vh;
  }
  .login-card {
    padding: 20px 18px 16px;
    border-radius: 12px;
  }
  .login-card__title {
    font-size: 18px;
  }
}
</style>
