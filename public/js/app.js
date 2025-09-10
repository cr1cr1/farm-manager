// Minimal progressive enhancement bootstrap for DataStar/DatastarUI + CSRF header

(function () {
  "use strict";

  function getCookie(name) {
    const m = document.cookie.match(new RegExp("(^|; )" + name.replace(/([.$?*|{}()[\]\\/+^])/g, "\\$1") + "=([^;]*)"));
    return m ? decodeURIComponent(m[2]) : "";
  }

  // Attach CSRF header to all non-idempotent fetches.
  const CSRF_COOKIE = (window.CSRF_COOKIE_NAME || "csrf_token");
  const CSRF_HEADER = (window.CSRF_HEADER_NAME || "X-CSRF-Token");

  const originalFetch = window.fetch;
  window.fetch = function (input, init) {
    init = init || {};
    const method = (init.method || (typeof input === "object" && input.method) || "GET").toUpperCase();
    if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS" && method !== "TRACE") {
      const headers = new Headers(init.headers || (typeof input === "object" ? input.headers : undefined) || {});
      if (!headers.has(CSRF_HEADER)) {
        const token = getCookie(CSRF_COOKIE);
        if (token) headers.set(CSRF_HEADER, token);
      }
      init.headers = headers;
    }
    return originalFetch.call(this, input, init);
  };

  // Initialize DataStar / DatastarUI if present.
  function initDataStar() {
    try {
      if (window.datastar && typeof window.datastar.init === "function") {
        window.datastar.init();
      }
      if (window.datastarui && typeof window.datastarui.init === "function") {
        window.datastarui.init();
      }
    } catch (e) {
      // no-op
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initDataStar);
  } else {
    initDataStar();
  }
})();
