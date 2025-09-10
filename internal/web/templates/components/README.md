# Templ Components (scaffold)

This directory will host reusable templ components for the hypermedia UI.

Planned components:

- Layout: App shell with header, footer, navigation, and content slots.
- Header: App title and user menu (logout form).
- Footer: Copyright and version.
- Navigation: Links for Dashboard (active state) and future sections.
- Forms: Input, PasswordInput, FormRow, FieldError, HiddenCsrfField.
- Buttons: Primary, Secondary, Danger, IconButton.
- Alerts: Info, Success, Warning, Danger.
- Notifications: Toast container (DataStar-driven).
- Modals: Modal and ConfirmDialog.

Notes:

- Components are designed for progressive enhancement: pages render without JS, DataStar/DatastarUI add interactivity via attributes (e.g., data-star-get).
- Keep props simple; prefer server-rendered HTML fragments over REST/JSON.
- Import CDN assets from the page/layout; avoid Node build steps.
