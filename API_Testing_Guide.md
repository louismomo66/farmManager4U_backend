# Farm Manager 4U API Testing Guide

## Base URL
```
http://localhost:9005
```

## Quick Test Sequence

### 1. Health Check
```bash
GET http://localhost:9005/health
```

### 2. User Signup
```bash
POST http://localhost:9005/api/auth/signup
Content-Type: application/json

{
  "firstName": "John",
  "lastName": "Doe",
  "email": "john.doe@example.com",
  "password": "password123",
  "role": "Farmer",
  "phoneNumber": "+1234567890",
  "address": "123 Farm Street, Farm City, FC 12345"
}
```

### 3. User Login
```bash
POST http://localhost:9005/api/auth/login
Content-Type: application/json

{
  "email": "john.doe@example.com",
  "password": "password123"
}
```
**Save the token from response for subsequent requests**

### 4. Create Farm
```bash
POST http://localhost:9005/api/farms
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json

{
  "name": "Sunshine Farm",
  "description": "A beautiful organic farm",
  "location": "123 Farm Road, Green Valley, CA 90210",
  "size": 50.5,
  "farmType": "Mixed",
  "status": "Active"
}
```
**Save the farmId from response**

### 5. Create Crop
```bash
POST http://localhost:9005/api/crops?farmId=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json

{
  "name": "Tomatoes",
  "plantingDate": "2024-03-15T00:00:00Z",
  "harvestDate": "2024-07-15T00:00:00Z",
  "quantity": 100,
  "status": "Growing",
  "notes": "Planted in greenhouse section A"
}
```

### 6. Create Livestock
```bash
POST http://localhost:9005/api/livestock?farmId=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json

{
  "type": "Cattle",
  "count": 25,
  "acquisitionDate": "2024-01-15T00:00:00Z",
  "healthStatus": "Healthy",
  "notes": "Angus cattle for beef production"
}
```

### 7. Create Employee
```bash
POST http://localhost:9005/api/employees?farmId=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json

{
  "firstName": "Jane",
  "lastName": "Smith",
  "position": "Farm Manager",
  "salary": 50000,
  "hireDate": "2024-01-01T00:00:00Z",
  "contactInfo": "jane.smith@example.com",
  "status": "Active"
}
```

## GET Requests

### Get All Farms
```bash
GET http://localhost:9005/api/farms
Authorization: Bearer YOUR_TOKEN_HERE
```

### Get Farm by ID
```bash
GET http://localhost:9005/api/farms?id=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
```

### Get Crops by Farm
```bash
GET http://localhost:9005/api/crops?farmId=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
```

### Get Livestock by Farm
```bash
GET http://localhost:9005/api/livestock?farmId=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
```

### Get Employees by Farm
```bash
GET http://localhost:9005/api/employees?farmId=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
```

## UPDATE Requests

### Update Farm
```bash
PUT http://localhost:9005/api/farms?id=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json

{
  "name": "Updated Farm Name",
  "description": "Updated description",
  "size": 55.0
}
```

### Update Crop
```bash
PUT http://localhost:9005/api/crops?id=YOUR_CROP_ID
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json

{
  "name": "Updated Crop Name",
  "quantity": 120,
  "status": "Harvested"
}
```

## DELETE Requests

### Delete Farm
```bash
DELETE http://localhost:9005/api/farms?id=YOUR_FARM_ID
Authorization: Bearer YOUR_TOKEN_HERE
```

### Delete Crop
```bash
DELETE http://localhost:9005/api/crops?id=YOUR_CROP_ID
Authorization: Bearer YOUR_TOKEN_HERE
```

## Testing Tips

1. **Start with Health Check** - Ensure the server is running
2. **Signup/Login First** - Get authentication token
3. **Create Farm** - Required for crops, livestock, and employees
4. **Test CRUD Operations** - Create, Read, Update, Delete for each entity
5. **Check Authorization** - Try requests without tokens to test security

## Postman Collection

Import the `Farm_Manager_4U_Postman_Collection.json` file into Postman for:
- Pre-configured requests
- Automatic token management
- Variable substitution
- Test scripts for response validation

## Common Response Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error
