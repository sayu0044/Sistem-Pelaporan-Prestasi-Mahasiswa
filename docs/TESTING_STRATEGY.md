# Testing Strategy

## 6. Testing Strategy

### 6.1 Unit Testing

Unit testing bertujuan untuk menguji setiap fungsi dan method secara individual dalam isolasi, dengan memastikan bahwa setiap unit kode bekerja sesuai yang diharapkan.

#### 6.1.1 Scope Unit Testing

**Fungsi dan Method yang Perlu Diuji:**

1. **Service Layer** (`app/service/`)
   - `AuthService`: Login, Register, ValidateToken, RefreshToken, GetProfile
   - `AchievementService`: Create, Update, Delete, Submit, Verify, Reject, GetAchievements, GetAchievementByID, GetAchievementHistory, UploadAttachment
   - `UserService`: CreateUser, UpdateUser, GetAllUsers, GetUserByID, DeleteUser, UpdateUserRole
   - `StudentService`: GetAllStudents, GetStudentByID, GetStudentAchievements, UpdateStudentAdvisor
   - `LecturerService`: GetAllLecturers, GetLecturerAdvisees
   - `ReportService`: GetStatistics, GetStudentStatistics

2. **Helper Functions** (`helper/`)
   - `TrimSpace()`
   - `IsEmpty()`
   - `Contains()`

3. **Middleware** (`middleware/`)
   - JWT validation functions
   - RBAC permission checking functions

#### 6.1.2 Mocking External Dependencies

**Dependencies yang Perlu Di-Mock:**

1. **Database Repositories**
   - Mock semua repository interfaces:
     - `UserRepository`
     - `RoleRepository`
     - `StudentRepository`
     - `LecturerRepository`
     - `AchievementRepository`
     - `AchievementHistoryRepository`
   - Lokasi mock: `test/app/repository/mocks.go`

2. **External Services**
   - JWT token generation/validation (gunakan secret key test)
   - Bcrypt password hashing (gunakan test password)
   - MongoDB operations (mock melalui repository)
   - File system operations (mock untuk upload attachment)

3. **Context**
   - Gunakan `context.Background()` untuk testing

#### 6.1.3 Testing Framework dan Tools

**Framework yang Digunakan:**
- Standard Go `testing` package
- `github.com/stretchr/testify` (opsional, untuk assertions yang lebih mudah)
- Mock structs manual (sudah ada di `test/app/repository/mocks.go`)

**Struktur Test File:**
- Naming convention: `*_test.go`
- Package: `service_test` untuk service tests
- Lokasi: `test/app/service/` (sudah ada)

#### 6.1.4 Test Structure dan Best Practices

**Struktur Test Function:**
```go
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange: Setup test data dan mocks
    // Act: Execute function yang diuji
    // Assert: Verify hasil
}
```

**Test Naming Convention:**
- `TestFunctionName_Success` - Test case sukses
- `TestFunctionName_ErrorCondition` - Test case error
- `TestFunctionName_EdgeCase` - Test case edge case

**Best Practices:**
1. Setiap test harus independent (tidak bergantung pada test lain)
2. Gunakan table-driven tests untuk multiple scenarios
3. Mock semua external dependencies
4. Test both success dan error cases
5. Test edge cases (nil values, empty strings, invalid inputs)
6. Test business logic validations

#### 6.1.5 Contoh Implementasi

**Contoh Test untuk Service Method:**

File: `test/app/service/user_service_test.go`
- Test `CreateUser_Success`
- Test `CreateUser_DuplicateUsername`
- Test `CreateUser_InvalidRole`
- Test `UpdateUser_Success`
- Test `UpdateUser_NotFound`
- Test `GetAllUsers_Success`
- Test `GetUserByID_Success`
- Test `GetUserByID_NotFound`
- Test `DeleteUser_Success`
- Test `DeleteUser_NotFound`

**Contoh Test untuk Helper Functions:**

File: `test/helper/helper_test.go`
- Test `TrimSpace()` dengan berbagai input
- Test `IsEmpty()` dengan empty dan non-empty strings
- Test `Contains()` dengan item yang ada dan tidak ada

#### 6.1.6 Coverage Goals

**Target Coverage:**
- Service layer: Minimal 80% code coverage
- Helper functions: 100% code coverage
- Critical business logic: 100% code coverage

**Prioritas Testing:**
1. High Priority: AuthService, AchievementService (core functionality)
2. Medium Priority: UserService, StudentService, LecturerService
3. Low Priority: Helper functions, utility methods

#### 6.1.7 Running Tests

**Command untuk menjalankan tests:**
```bash
# Run all tests
go test ./...

# Run tests dengan coverage
go test -cover ./...

# Run tests dengan verbose output
go test -v ./...

# Run specific test file
go test -v ./test/app/service/user_service_test.go

# Run specific test function
go test -v -run TestLogin_Success ./test/app/service/
```

#### 6.1.8 Mock Implementation Details

**Mock Repository Pattern:**
- Setiap mock repository memiliki fields untuk menyimpan expected return values
- Setiap mock repository memiliki error fields untuk simulate errors
- Mock methods return values sesuai dengan fields yang di-set

**Contoh Mock Structure (sudah ada di `test/app/repository/mocks.go`):**
```go
type MockUserRepository struct {
    ByID       *model.User
    ByUsername *model.User
    ByEmail    *model.User
    All        []model.User
    CreateErr  error
    FindErr    error
    UpdateErr  error
    DeleteErr  error
}
```

#### 6.1.9 Test Data Management

**Test Fixtures:**
- Gunakan UUID generator untuk test IDs
- Gunakan consistent test data untuk reproducibility
- Setup dan teardown jika diperlukan (untuk integration tests nanti)

**Test Constants:**
- JWT secret untuk testing: `"test-secret"`
- Test user credentials: username, password, email
- Test role IDs dan names

#### 6.1.10 Test Files yang Sudah Dibuat

**Service Tests:**
- `test/app/service/auth_service_test.go` - AuthService tests (sudah ada)
- `test/app/service/achievement_service_test.go` - AchievementService tests (sudah ada)
- `test/app/service/user_service_test.go` - UserService tests (baru dibuat)
- `test/app/service/student_service_test.go` - StudentService tests (baru dibuat)
- `test/app/service/lecturer_service_test.go` - LecturerService tests (baru dibuat)
- `test/app/service/report_service_test.go` - ReportService tests (baru dibuat)

**Helper Tests:**
- `test/helper/helper_test.go` - Helper functions tests (baru dibuat)

#### 6.1.11 Next Steps

1. Lengkapi unit tests untuk middleware functions
2. Tambahkan integration tests untuk end-to-end scenarios
3. Setup CI/CD untuk run tests automatically
4. Monitor test coverage dan improve coverage untuk critical paths
5. Tambahkan performance tests untuk critical operations

