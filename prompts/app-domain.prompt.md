Generate pages and templ templates along with the Go handler logic to manage the internal/domain objects for a bio turkey farm management system. The domain objects include Barn, FeedType, Staff, Flock, FeedingRecord, HealthCheck, MortalityRecord, ProductionBatch, SlaughterRecord, InventoryItem, Customer, Order, and OrderItem, fully respecting their relationships as defined in the data model.

You MUST use only datastar (https://github.com/starfederation/datastar) for the data management and state synchronization layers.

The UI components MUST be from datastarUI (https://github.com/coreycole/datastarui) for all standard UI elements.

If any domain objects or relationships require UI controls or data presentation patterns not covered by datastarUI components, create new UI components limited to those missing features.

New components MUST follow the datastarUI directory and code pattern, encapsulated in their own directories with clear names representing their functionality.

Maintain clear and concise component structure and idiomatic Go code in handlers and templ templates.

The templ templates should declare reusable form components and data grids or lists as appropriate to each domain entity, leveraging datastarUI components for inputs, tables, modals, and notifications.

Respect foreign key relationships for nested or connected data (e.g., show barn selection on Flock forms, link feeding records to flocks and feed types).

Provide CRUD handlers in Go for each domain entity, implementing proper fetching, creation, updating, and deletion with necessary validation.

Exclude user management and audit logging entities and logic from this generation.

Ensure templates and pages are structured to enable efficient navigation and data management across the farm management system.

Deliver instructions in a way that supports clean scalable code generation using templ and integrates seamlessly with datastar and datastarUI tools.
