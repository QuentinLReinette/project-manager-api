# API Endpoint Documentation

All endpoints (except Public Authentication routes) are protected and require a valid JWT token. The API operates using secure `HttpOnly` cookies. Once authenticated, the browser automatically attaches the cookie to all requests.

For development/testing via tools like `curl` or Postman, you can pass the session token using the `Cookie` header:
`Cookie: token={your_jwt_token}`

---

## CORS Support

The API has full CORS support enabled. Preflight `OPTIONS` requests are handled and return a `204 No Content` status with the following CORS headers:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

---

## Summary of Endpoints

| Category     | Method | Endpoint                          | Description                                          | Auth Required |
| ------------ | ------ | --------------------------------- | ---------------------------------------------------- | ------------- |
| **Health**   | GET    | `/ping`                           | Base API sanity check                                | No            |
| **Auth**     | POST   | `/api/auth/register`              | Register a new user                                  | No            |
| **Auth**     | POST   | `/api/auth/login`                 | Log in and set `token` cookie                        | No            |
| **Auth**     | GET    | `/api/auth/me`                    | Get authenticated user profile                       | Yes           |
| **Users**    | GET    | `/api/users?q={query}`            | Search users by name/email (min 3 chars)             | Yes           |
| **Projects** | GET    | `/api/projects`                   | List all projects (owned or participated)            | Yes           |
| **Projects** | POST   | `/api/projects`                   | Create a new project workspace                       | Yes           |
| **Projects** | GET    | `/api/projects/{id}`              | Get detailed project (incl. tasks/members)           | Yes           |
| **Projects** | PUT    | `/api/projects/{id}`              | Update project title/description                     | Yes (Owner)   |
| **Projects** | DELETE | `/api/projects/{id}`              | Delete a project workspace                           | Yes (Owner)   |
| **Projects** | POST   | `/api/projects/{id}/participants` | Add a participant by registered email                | Yes (Owner)   |
| **Tasks**    | GET    | `/api/tasks?project_id={id}`      | List all tasks for a project (incl. filter `status`) | Yes (Member)  |
| **Tasks**    | POST   | `/api/tasks`                      | Create a new task under a project                    | Yes (Member)  |
| **Tasks**    | GET    | `/api/tasks/{id}`                 | Get detailed task info (incl. assignee)              | Yes (Member)  |
| **Tasks**    | PUT    | `/api/tasks/{id}`                 | Update task title, status, or assignment             | Yes (Member)  |
| **Tasks**    | DELETE | `/api/tasks/{id}`                 | Delete a task                                        | Yes (Member)  |

---

## Detailed Payloads and Responses

### 0. Health Check

- **Sanity check (`GET /ping`)**
  - **Description:** Returns the status of the API engine. Does not require authentication.
  - **Response (200 OK):**

    ```json
    {
      "message": "pong",
      "status": "running"
    }
    ```

---

### 1. Authentication

- **Register a new account (`POST /api/auth/register`)**
  - **Description:** Registers a new user inside the application, issues a session JWT, and sets it in an HTTP-only `token` cookie.
  - **Payload:**

    ```json
    {
      "name": "Alice Ynov",
      "email": "alice@ynov.com",
      "password": "strongpassword"
    }
    ```

  - **Response (201 Created):**

    ```json
    {
      "id": 6,
      "name": "Alice Ynov",
      "email": "alice@ynov.com"
    }
    ```

  - **Error Responses:**
    - `400 Bad Request`: Payload is invalid.
    - `409 Conflict`: Email is already registered.
    - `422 Unprocessable Entity`: Required fields (`name`, `email`, `password`) are missing.

- **Login (`POST /api/auth/login`)**
  - **Description:** Authenticates a user, sets the session JWT in an HTTP-only `token` cookie, and returns the token in the JSON body response.
  - **Payload:**

    ```json
    {
      "email": "alice@ynov.com",
      "password": "strongpassword"
    }
    ```

  - **Response (200 OK):**

    ```json
    {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzk1MzA0MTMsInVzZXJfaWQiOjZ9.4tc-e...",
      "user": {
        "id": 6,
        "name": "Alice Ynov",
        "email": "alice@ynov.com"
      }
    }
    ```

  - **Error Responses:**
    - `401 Unauthorized`: Invalid email or password.

- **Who Am I (`GET /api/auth/me`)**
  - **Description:** Returns the currently authenticated user's details. Requires authentication (valid `token` cookie).
  - **Response (200 OK):**

    ```json
    {
      "id": 6,
      "name": "Alice Ynov",
      "email": "alice@ynov.com"
    }
    ```

  - **Error Responses:**
    - `401 Unauthorized`: Authentication cookie missing, expired, or invalid.

---

### 2. Users (Protected)

- **Search Users (`GET /api/users?q={query}`)**
  - **Description:** Searches for registered users matching the query `q` (against `name` or `email`). Used by the frontend autocomplete to add participants to a project.
  - **Parameters:**
    - `q` (string, required): The search keyword. Must be at least **3 characters** long.
  - **Response (200 OK):**

    ```json
    [
      {
        "id": 7,
        "name": "Bob Martin",
        "email": "bob@ynov.com"
      }
    ]
    ```

  - **Error Responses:**
    - `400 Bad Request`: Query parameter `q` is missing or is less than 3 characters long.

---

### 3. Projects (Protected)

- **List Projects (`GET /api/projects`)**
  - **Description:** Returns all projects where the authenticated user is either the Owner or an active Participant.
  - **Response (200 OK):**

    ```json
    [
      {
        "id": 2,
        "title": "Marketing Redesign",
        "description": "Redesign of the company web portal",
        "owner_id": 6,
        "owner": {
          "id": 6,
          "name": "Alice Ynov",
          "email": "alice@ynov.com"
        },
        "created_at": "2026-05-22T10:07:19Z",
        "updated_at": "2026-05-22T10:07:25Z"
      }
    ]
    ```

- **Create Project (`POST /api/projects`)**
  - **Description:** Creates a new project workspace owned by the authenticated user.
  - **Payload:**

    ```json
    {
      "title": "Marketing Redesign",
      "description": "Redesign of the company web portal"
    }
    ```

  - **Response (201 Created):**

    ```json
    {
      "id": 2,
      "title": "Marketing Redesign",
      "description": "Redesign of the company web portal",
      "owner_id": 6,
      "created_at": "2026-05-22T10:07:19Z",
      "updated_at": "2026-05-22T10:07:19Z"
    }
    ```

- **Get Project Details (`GET /api/projects/{id}`)**
  - **Description:** Returns detailed info of a single project, including owner profile, list of participants, and all tasks.
  - **Access Control:** User must be the Owner or an active Participant of this project.
  - **Error Responses:**
    - `400 Bad Request`: Project identifier is invalid.
    - `403 Forbidden`: User is not the owner or a participant of this project.
    - `404 Not Found`: Project workspace target not found.
  - **Response (200 OK):**

    ```json
    {
      "id": 2,
      "title": "Marketing Redesign",
      "description": "Redesign of the company web portal",
      "owner_id": 6,
      "owner": {
        "id": 6,
        "name": "Alice Ynov",
        "email": "alice@ynov.com"
      },
      "participants": [
        {
          "id": 7,
          "name": "Bob Martin",
          "email": "bob@ynov.com"
        }
      ],
      "tasks": [
        {
          "id": 2,
          "title": "Setup Database",
          "description": "Create schema and seed tables",
          "status": "in_progress",
          "project_id": 2,
          "assigned_to_id": 7
        }
      ],
      "created_at": "2026-05-22T10:07:19Z",
      "updated_at": "2026-05-22T10:07:25Z"
    }
    ```

- **Update Project (`PUT /api/projects/{id}`)**
  - **Description:** Updates the project workspace details.
  - **Access Control:** Restricted to the project Owner.
  - **Payload:**

    ```json
    {
      "title": "Updated Marketing Redesign",
      "description": "Updated description"
    }
    ```

  - **Error Responses:**
    - `400 Bad Request`: Project identifier is invalid, or modification updates payload JSON is invalid.
    - `403 Forbidden`: Access denied (user is not the project Owner).
    - `404 Not Found`: Project workspace target not found.
  - **Response (200 OK):** Returns the updated project object.

- **Delete Project (`DELETE /api/projects/{id}`)**
  - **Description:** Deletes the project workspace and all its tasks (via Cascade).
  - **Access Control:** Restricted to the project Owner.
  - **Error Responses:**
    - `400 Bad Request`: Project identifier is invalid.
    - `403 Forbidden`: Access denied (user is not the project Owner).
    - `404 Not Found`: Project workspace target not found.
  - **Response (200 OK):**

    ```json
    {
      "message": "Project workspace wiped successfully"
    }
    ```

- **Add Participant (`POST /api/projects/{id}/participants`)**
  - **Description:** Adds a user to the project as a participant by their email.
  - **Access Control:** Restricted to the project Owner.
  - **Payload:**

    ```json
    {
      "email": "bob@ynov.com"
    }
    ```

  - **Response (200 OK):**

    ```json
    {
      "message": "Participant successfully attached to project"
    }
    ```

  - **Error Responses:**
    - `400 Bad Request`: Project identifier is invalid, or payload JSON is invalid / missing required fields (`email`).
    - `403 Forbidden`: Access denied (user is not the project Owner).
    - `404 Not Found`: Project workspace target not found, or the email address is not registered inside the application database.

---

### 4. Tasks (Protected)

- **List Project Tasks (`GET /api/tasks?project_id={id}&status={filter}`)**
  - **Description:** Lists all tasks of a project. Can be filtered by status.
  - **Parameters:**
    - `project_id` (integer, required): ID of the project.
    - `status` (string, optional): Filter tasks by state. Allowed values: `todo`, `in_progress`, `done`.
  - **Access Control:** User must be the Owner or a Participant of the project.
  - **Error Responses:**
    - `400 Bad Request`: Missing `project_id` query parameter, or query `status` parameter is invalid (not one of `todo`, `in_progress`, `done`).
    - `403 Forbidden`: User is not a participant or owner of the specified project workspace.
  - **Response (200 OK):**

    ```json
    [
      {
        "id": 2,
        "title": "Setup Database",
        "description": "Create schema and seed tables",
        "status": "in_progress",
        "project_id": 2,
        "assigned_to_id": 7,
        "assigned_to": {
          "id": 7,
          "name": "Bob Martin",
          "email": "bob@ynov.com"
        }
      }
    ]
    ```

- **Create Task (`POST /api/tasks`)**
  - **Description:** Creates a new task explicitly linked to a project. Newly created tasks start with the `"todo"` status by default.
  - **Access Control:** User must be the Owner or a Participant of the project.
  - **Payload:**

    ```json
    {
      "title": "Setup Database",
      "description": "Create schema and seed tables",
      "project_id": 2
    }
    ```

  - **Error Responses:**
    - `400 Bad Request`: Payload is invalid, or missing required fields (`title`, `project_id`).
    - `403 Forbidden`: User is not a participant or owner of the specified project workspace.
  - **Response (201 Created):** Returns the created task object.

- **Get Task Details (`GET /api/tasks/{id}`)**
  - **Description:** Returns detailed info of a single task, preloading the assigned user details.
  - **Access Control:** User must be the Owner or a Participant of the project containing the task.
  - **Error Responses:**
    - `400 Bad Request`: Task identifier is invalid.
    - `403 Forbidden`: User is not a participant or owner of the project containing the task.
    - `404 Not Found`: Task not found.
  - **Response (200 OK):**

    ```json
    {
      "id": 2,
      "title": "Setup Database",
      "description": "Create schema and seed tables",
      "status": "in_progress",
      "project_id": 2,
      "assigned_to_id": 7,
      "assigned_to": {
        "id": 7,
        "name": "Bob Martin",
        "email": "bob@ynov.com"
      },
      "created_at": "2026-05-22T10:10:40Z",
      "updated_at": "2026-05-22T10:11:38Z"
    }
    ```

- **Update Task (`PUT /api/tasks/{id}`)**
  - **Description:** Updates the task's properties. Can be used to change the task status or assign/reassign it to a participant.
  - **Access Control:** User must be a member of the project workspace.
  - **Payload:**

    ```json
    {
      "title": "New Title",
      "description": "New description",
      "status": "in_progress",
      "assigned_to_id": 7
    }
    ```

  - **Error Responses:**
    - `400 Bad Request`: Task identifier is invalid, or payload JSON is invalid.
    - `403 Forbidden`: User is not a member of the project workspace.
    - `404 Not Found`: Task not found.
    - `422 Unprocessable Entity`: Provided fields failed validation:
      - `status`: Must be one of `todo`, `in_progress`, or `done`.
      - `assigned_to_id`: Must refer to a registered user who is currently the project Owner or a Participant.
  - **Response (200 OK):** Returns the updated task object.

- **Delete Task (`DELETE /api/tasks/{id}`)**
  - **Description:** Deletes the task.
  - **Access Control:** User must be a member of the project workspace.
  - **Error Responses:**
    - `400 Bad Request`: Task identifier is invalid.
    - `403 Forbidden`: User is not a member of the project workspace.
    - `404 Not Found`: Task not found.
  - **Response (200 OK):**

    ```json
    {
      "message": "Task deleted successfully"
    }
    ```
