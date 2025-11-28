package permissions

// Permission bits (using bit flags)
const (
	// Basic permissions
	PermissionViewReports   = 1 << 0 // 1 - View reports
	PermissionAddReports    = 1 << 1 // 2 - Add reports
	PermissionEditReports   = 1 << 2 // 4 - Edit reports
	PermissionDeleteReports = 1 << 3 // 8 - Delete reports

	// Taxi management
	PermissionViewTaxis   = 1 << 4 // 16 - View taxis
	PermissionAddTaxis    = 1 << 5 // 32 - Add taxis
	PermissionEditTaxis   = 1 << 6 // 64 - Edit taxis
	PermissionDeleteTaxis = 1 << 7 // 128 - Delete taxis

	// Expense management
	PermissionViewExpenses   = 1 << 8  // 256 - View expenses
	PermissionAddExpenses    = 1 << 9  // 512 - Add expenses
	PermissionEditExpenses   = 1 << 10 // 1024 - Edit expenses
	PermissionDeleteExpenses = 1 << 11 // 2048 - Delete expenses

	// Deposit management
	PermissionViewDeposits   = 1 << 12 // 4096 - View deposits
	PermissionAddDeposits    = 1 << 13 // 8192 - Add deposits
	PermissionEditDeposits   = 1 << 14 // 16384 - Edit deposits
	PermissionDeleteDeposits = 1 << 15 // 32768 - Delete deposits

	// User management
	PermissionViewUsers   = 1 << 16 // 65536 - View users
	PermissionAddUsers    = 1 << 17 // 131072 - Add users
	PermissionEditUsers   = 1 << 18 // 262144 - Edit users
	PermissionDeleteUsers = 1 << 19 // 524288 - Delete users

	// Tenant management (admin only)
	PermissionManageTenants = 1 << 20 // 1048576 - Manage tenants
)

// Permission masks for roles
// These values can be overridden via environment variables
var (
	// Admin has all permissions (default: 0xFFFFFFFF, can be set via JWT_ADMIN_PERMISSION_MASK)
	PermissionAdmin = 0xFFFFFFFF // All bits set

	// Owner has all permissions for their business (default: 0xFFFFF, can be set via JWT_OWNER_PERMISSION_MASK)
	PermissionOwner = PermissionViewReports | PermissionAddReports | PermissionEditReports | PermissionDeleteReports |
		PermissionViewTaxis | PermissionAddTaxis | PermissionEditTaxis | PermissionDeleteTaxis |
		PermissionViewExpenses | PermissionAddExpenses | PermissionEditExpenses | PermissionDeleteExpenses |
		PermissionViewDeposits | PermissionAddDeposits | PermissionEditDeposits | PermissionDeleteDeposits

	// Manager can only add reports (default: 0x3 = View + Add, can be set via JWT_MANAGER_PERMISSION_MASK)
	PermissionManager = PermissionViewReports | PermissionAddReports | PermissionEditReports |
		PermissionViewDeposits | PermissionAddDeposits | PermissionEditDeposits |
		PermissionViewTaxis

	// Mechanic can view and manage taxis (default: 0x70, can be set via JWT_MECHANIC_PERMISSION_MASK)
	PermissionMechanic = PermissionViewTaxis | PermissionViewReports

	// Driver can view and add reports (default: 0x3, can be set via JWT_DRIVER_PERMISSION_MASK)
	PermissionDriver = PermissionViewReports | PermissionAddReports
)

// SetPermissionMasks allows setting permission masks from config
func SetPermissionMasks(admin, owner, manager, mechanic, driver int) {
	PermissionAdmin = admin
	PermissionOwner = owner
	PermissionManager = manager
	PermissionMechanic = mechanic
	PermissionDriver = driver
}

// HasPermission checks if the user has a specific permission
func HasPermission(userPermission int, requiredPermission int) bool {
	// Admin has all permissions
	// Check for -1 (all bits set in signed int, used when stored in PostgreSQL INTEGER)
	// or if it matches PermissionAdmin (which may be 0xFFFFFFFF on 64-bit systems)
	if userPermission == -1 || userPermission == PermissionAdmin {
		return true
	}
	// Also check bitwise: if all bits are set, user has all permissions
	if userPermission&PermissionAdmin == PermissionAdmin {
		return true
	}
	return userPermission&requiredPermission != 0
}

// HasAnyPermission checks if the user has any of the required permissions
func HasAnyPermission(userPermission int, requiredPermissions ...int) bool {
	// Admin has all permissions
	if userPermission == -1 || userPermission == PermissionAdmin {
		return true
	}
	// Also check bitwise: if all bits are set, user has all permissions
	if userPermission&PermissionAdmin == PermissionAdmin {
		return true
	}
	for _, perm := range requiredPermissions {
		if userPermission&perm != 0 {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the user has all required permissions
func HasAllPermissions(userPermission int, requiredPermissions ...int) bool {
	// Admin has all permissions
	if userPermission == -1 || userPermission == PermissionAdmin {
		return true
	}
	// Also check bitwise: if all bits are set, user has all permissions
	if userPermission&PermissionAdmin == PermissionAdmin {
		return true
	}
	for _, perm := range requiredPermissions {
		if userPermission&perm == 0 {
			return false
		}
	}
	return true
}

// GetPermissionForRole returns the permission mask for a role name (for backward compatibility)
func GetPermissionForRole(role string) int {
	switch role {
	case "admin":
		return PermissionAdmin
	case "owner":
		return PermissionOwner
	case "manager":
		return PermissionManager
	case "mechanic":
		return PermissionMechanic
	case "driver":
		return PermissionDriver
	default:
		return PermissionDriver // Default to driver permissions
	}
}

// GetRoleName returns the role name for a permission mask (for display purposes)
func GetRoleName(permission int) string {
	// Check for admin: -1 (all bits set in signed int) or matches PermissionAdmin
	if permission == -1 || permission == PermissionAdmin || (permission&PermissionAdmin != 0 && permission&PermissionAdmin == PermissionAdmin) {
		return "admin"
	}
	if permission == PermissionOwner {
		return "owner"
	}
	if permission == PermissionManager {
		return "manager"
	}
	if permission == PermissionMechanic {
		return "mechanic"
	}
	if permission == PermissionDriver {
		return "driver"
	}
	return "custom"
}
