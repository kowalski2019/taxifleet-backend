# Seed Data Documentation

## Default Tenant and Users

The application includes seed data for testing purposes. This data is created using the Go seed script: `cmd/seed/main.go`

### Default Tenant

- **Name**: Gnakpa Transport
- **Subdomain**: gnakpa-transport

### Default Users

All users have the password pattern: `{role}123`

#### Owner
- **Email**: tanguy.gnakpa@gnakpa-transport.com
- **Password**: `owner123`
- **Name**: Tanguy Gnakpa
- **Role**: owner
- **Phone**: +1234567890

#### Manager
- **Email**: manager@gnakpa-transport.com
- **Password**: `manager123`
- **Name**: Gnakpa Sister
- **Role**: manager
- **Phone**: +1234567891

#### Mechanic
- **Email**: mechanic@gnakpa-transport.com
- **Password**: `mechanic123`
- **Name**: Just Mechanicer
- **Role**: mechanic
- **Phone**: +1234567892

#### Driver
- **Email**: driver@gnakpa-transport.com
- **Password**: `driver123`
- **Name**: Test Driver
- **Role**: driver
- **Phone**: +1234567893

### Test Taxis

Three test taxis are created:

1. **ABC-123** - Toyota Camry 2020 (White) - Active
2. **XYZ-789** - Honda Accord 2021 (Black) - Active
3. **DEF-456** - Nissan Altima 2019 (Silver) - Maintenance

## Usage

To create the seed data, run:

```bash
cd backend
go run cmd/seed/main.go
```

The script will:
- Check if the tenant already exists (skips if found)
- Create the tenant "Gnakpa Transport"
- Create all users with properly hashed passwords
- Create 3 test taxis

To reset the seed data, you can:
1. Delete the tenant and related data from the database
2. Re-run the seed script

The seed script is idempotent - it won't create duplicates if the tenant already exists.

## Security Note

⚠️ **These are test credentials only!** Do not use these passwords in production. The seed data should be removed or modified before deploying to production.

