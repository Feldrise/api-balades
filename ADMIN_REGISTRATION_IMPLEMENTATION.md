# Admin Registration Management Implementation

## Summary

Successfully implemented Phases 1, 2, and 3 of the admin registration management system for the Balade Écologique API.

## Phase 1: Admin Permissions & Routes ✅

### 1. Permission Middleware (`pkg/authentication/permissions.go`)

- Added `RequirePermission()` middleware for single permission checks
- Added `RequireAnyPermission()` middleware for multiple permission checks
- Added `RequireAuthentication()` middleware for basic auth checks
- Added `ErrForbidden()` to error handling system

### 2. Admin Permissions (`database/seed/seedv4.go`)

Created 5 new admin permissions:

- `manage:registration` - Full CRUD on registrations
- `view:all-registrations` - Read all registrations
- `update:registration-status` - Status changes
- `update:registration-details` - Edit registration info
- `bulk:registration-actions` - Bulk operations

### 3. Updated Routes (`pkg/registration/routes.go`)

Added admin-specific route group:

- `GET /admin` - List all registrations with filters
- `PUT /admin/{id}` - Update registration details
- `PUT /admin/{id}/status` - Update registration status
- `DELETE /admin/{id}` - Delete registration
- `POST /admin/bulk-action` - Bulk operations

## Phase 2: Enhanced Filtering & Search ✅

### 1. Enhanced Database Filter (`database/dbmodel/rambleregistration.go`)

Extended `RambleRegistrationFilter` with:

- Date range filtering (`DateFrom`, `DateTo`)
- Full-text search in name/email (`Search`)
- Ramble title search (`RambleTitle`)
- Multiple status filtering (`Statuses`)
- Pagination (`Limit`, `Offset`)
- Sorting (`SortBy`, `SortOrder`)

### 2. Advanced Query Building

- Support for ILIKE searches (case-insensitive)
- JOIN with rambles table for title search
- Proper SQL injection protection
- Flexible sorting on multiple fields

### 3. Count Method (`CountAll`)

- Added pagination support with total count
- Consistent filtering between count and data queries

## Phase 3: Admin Update Capabilities ✅

### 1. Admin Models (`pkg/model/admin_registration.go`)

Created comprehensive admin-specific payloads:

- `AdminRegistrationUpdatePayload` - Update registration details
- `AdminRegistrationStatusUpdatePayload` - Status updates with reason
- `BulkRegistrationActionPayload` - Bulk operations
- `AdminRegistrationFilterPayload` - Advanced filtering
- `AdminRegistrationListResponse` - Paginated responses
- `BulkActionResult` - Bulk operation results

### 2. Admin Controller (`pkg/registration/admin_controller.go`)

Implemented 5 admin endpoints:

#### `AdminGetAllRegistrations`

- Advanced filtering and search
- Pagination with total count
- Permission: `view:all-registrations`

#### `AdminUpdateRegistration`

- Update name, email, phone, status
- Maintains business logic for status changes
- Permission: `update:registration-details`

#### `AdminUpdateRegistrationStatus`

- Status updates with optional reason
- Email notification support
- Automatic waiting list promotion
- Permission: `update:registration-status`

#### `AdminDeleteRegistration`

- Safe deletion with waiting list promotion
- Permission: `manage:registration`

#### `AdminBulkRegistrationAction`

- Bulk confirm/cancel/move to waiting/delete
- Individual error tracking
- Email notifications
- Permission: `bulk:registration-actions`

### 3. Enhanced Existing Endpoints

- Updated `GetRambleRegistrations` with proper permission checks
- Added `view:all-registrations` permission requirement

## API Endpoints Summary

### Admin Registration Management

| Method | Endpoint                           | Permission                    | Description                                    |
| ------ | ---------------------------------- | ----------------------------- | ---------------------------------------------- |
| GET    | `/registrations/admin`             | `view:all-registrations`      | List all registrations with advanced filtering |
| PUT    | `/registrations/admin/{id}`        | `update:registration-details` | Update registration details                    |
| PUT    | `/registrations/admin/{id}/status` | `update:registration-status`  | Update registration status                     |
| DELETE | `/registrations/admin/{id}`        | `manage:registration`         | Delete registration                            |
| POST   | `/registrations/admin/bulk-action` | `bulk:registration-actions`   | Bulk operations                                |

### Query Parameters for GET /registrations/admin

- `ramble_id` - Filter by ramble ID
- `user_id` - Filter by user ID
- `email` - Filter by exact email
- `status` - Filter by single status
- `statuses` - Filter by multiple statuses (array)
- `date_from` - Filter registrations from date (YYYY-MM-DD)
- `date_to` - Filter registrations to date (YYYY-MM-DD)
- `search` - Search in name and email
- `ramble_title` - Search in ramble title
- `page` - Page number (starts at 1)
- `per_page` - Items per page (max 500)
- `sort_by` - Sort field (created_at, registration_date, first_name, last_name, email, status)
- `sort_order` - Sort order (asc, desc)

## Features Implemented

### ✅ Completed Features

1. **Full Admin Permission System** - Role-based access control
2. **Advanced Filtering** - Date ranges, search, multiple criteria
3. **Pagination** - Efficient large dataset handling
4. **Bulk Operations** - Mass status updates and deletions
5. **Status Management** - Proper business logic for status transitions
6. **Waiting List Automation** - Automatic promotion when spots open
7. **Error Handling** - Comprehensive error responses
8. **Input Validation** - Request payload validation
9. **Database Optimization** - Efficient queries with preloading

### 🔧 Technical Improvements

1. **Permission Middleware** - Reusable authentication/authorization
2. **Repository Pattern** - Consistent data access
3. **Model Separation** - Admin vs user models
4. **Query Building** - Dynamic, safe SQL generation
5. **Error Aggregation** - Bulk operation error tracking

## Database Changes

### New Permissions Added

```sql
INSERT INTO permissions (id, name, readable_name, description) VALUES
(16, 'manage:registration', 'Gérer les inscriptions', 'Permet de gérer toutes les inscriptions (lecture, modification, suppression)'),
(17, 'view:all-registrations', 'Voir toutes les inscriptions', 'Permet de voir toutes les inscriptions de toutes les balades'),
(18, 'update:registration-status', 'Modifier le statut des inscriptions', 'Permet de modifier le statut des inscriptions (confirmer, annuler, etc.)'),
(19, 'update:registration-details', 'Modifier les détails d\'inscription', 'Permet de modifier les informations personnelles des inscriptions'),
(20, 'bulk:registration-actions', 'Actions en lot sur les inscriptions', 'Permet d\'effectuer des actions en lot sur plusieurs inscriptions');
```

## Usage Examples

### Filter registrations by status and date range

```bash
GET /registrations/admin?status=pending&date_from=2025-01-01&date_to=2025-01-31&page=1&per_page=20
```

### Search registrations

```bash
GET /registrations/admin?search=john.doe@example.com&ramble_title=Nature Walk
```

### Update registration status

```json
PUT /registrations/admin/123/status
{
  "status": "confirmed",
  "reason": "Manual approval by admin",
  "send_email": true
}
```

### Bulk cancel registrations

```json
POST /registrations/admin/bulk-action
{
  "registration_ids": [1, 2, 3, 4],
  "action": "cancel",
  "reason": "Event cancelled due to weather",
  "send_email": true
}
```

## Next Steps (Future Enhancements)

1. **Email Templates** - Admin-specific notification emails
2. **Audit Logging** - Track admin actions for compliance
3. **Export Functionality** - CSV/Excel export of registration data
4. **Dashboard Analytics** - Registration statistics and charts
5. **Advanced Permissions** - Resource-specific permissions (per ramble)
6. **Batch Import** - CSV import for bulk registrations

## Security Considerations

- ✅ Permission-based access control
- ✅ Input validation and sanitization
- ✅ SQL injection prevention
- ✅ Authentication required for all admin actions
- ✅ Error messages don't leak sensitive information

The implementation provides a robust, secure, and scalable admin interface for managing registrations with comprehensive filtering, search, and bulk operation capabilities.
