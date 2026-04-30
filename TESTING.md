### **Module: Authentication & Authorization**

| ID | Test Scenario | Test Steps | Expected Result | Status |
| :--- | :--- | :--- | :--- | :--- |
| **AUTH-01** | Successful Registration | Submit valid username, email, and password. | 201 Created, user record created in DB. | [OK] |
| **AUTH-02** | Duplicate Registration | Register using an existing email or username. | 409 Conflict, error message: "Email/Username already exists". | [OK] |
| **AUTH-03** | Input Validation | Submit empty fields or invalid email format. | 422 Unprocessable Entity, validation error details | [OK] |
| **AUTH-04** | Successful Login | Submit credentials of a verified account. | 200 OK, returns Access Token & Refresh Token. | [OK] |
| **AUTH-05** | Login Unverified Account | Submit credentials of an unverified account. | 401 Unauthorized, error: "Please verify your email". | [OK] |
| **AUTH-06** | Email Verification | Access verification link with a valid token. | 200 OK, `is_verified` status updated to `true`. | [OK] |
| **AUTH-07** | Resend Email Verification | Send new verification link with a valid token. | 200 OK, New link send to email successfully. | [OK] |
| **AUTH-08** | Resend Email Verification to Verified Account | Send new verification link to verified account. | 409 Conflict, error message: "Account already verified". | [OK] |
| **AUTH-09** | Successful Token Refresh | Submit a valid Refresh Token to `/refresh`. | 200 OK, returns a new Token Pair (Rotation). | [OK] |