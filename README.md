# Travel-Collab

Travel-Collab is a web service for collaborative trip planning. It combines trip members, routes, map points, shared expenses and real-time chat in one workspace.

Live demo: **https://travel-collab.ru**

## Purpose

The service is designed for small travel groups that need a shared place to organize a trip: invite participants, plan a route, discuss details and track shared expenses.

Typical use cases:

- planning a group trip with several participants;
- keeping route points and visit order in one place;
- discussing trip details in a shared chat;
- recording group expenses and seeing who owes whom;
- separating access rights between owner, editor and viewer roles.

## Main features

### Trips and members

- user registration and authentication;
- trip creation;
- invite-based trip joining;
- member list;
- role-based access model: `owner`, `editor`, `viewer`.

### Routes and map

- map-based trip planning;
- route entities inside a trip;
- ordered locations for each route;
- visual connection of route points;
- manual point ordering.

The route line is used as a visual representation of the planned visit order. It is not a turn-by-turn navigation route.

### Shared expenses

- expense creation inside a trip;
- linking expenses to a trip, route or location;
- several payers and participants per expense;
- equal split and manual split modes;
- balance calculation between trip members;
- simplified settlement suggestions.

### Real-time chat

- trip-level chat;
- WebSocket-based real-time message delivery;
- messages are isolated by trip;
- only trip members can connect to the trip chat.

## Technology stack

### Backend

- Go
- Chi router
- pgx
- PostgreSQL
- PostGIS
- Gorilla WebSocket
- JWT authentication
- bcrypt password hashing

### Frontend

- Vue 3
- TypeScript
- Vite
- Pinia
- Axios
- Leaflet / OpenStreetMap
- Capacitor for Android build

### Infrastructure

- Docker
- Docker Compose
- Caddy reverse proxy
- GitHub Actions

## High-level architecture

```text
Browser / Android WebView
        |
        | REST API + WebSocket
        v
Go backend
        |
        v
PostgreSQL + PostGIS
```

The frontend communicates with the backend through REST endpoints and a WebSocket connection. The backend contains the main business logic, validates user access and works with the database. PostgreSQL stores application data, while PostGIS is used for map-related location data.

## Backend structure

The backend is organized around several main layers:

- configuration loading and validation;
- HTTP API handlers;
- authentication and authorization middleware;
- repository layer for database operations;
- WebSocket hub for trip chat;
- domain models for users, trips, members, routes, locations, expenses and messages.

Main backend domains:

- users and authentication;
- trips and membership;
- routes and locations;
- expenses and balances;
- messages and WebSocket chat.

## Frontend structure

The frontend is built as a Vue single-page application.

Main interface areas:

- authentication screens;
- trip list;
- trip workspace;
- map and itinerary panels;
- members panel;
- expenses panel;
- chat panel.

State management is handled through Pinia stores. API calls are placed in a separate HTTP layer. Real-time trip updates are handled through a dedicated WebSocket composable.

## Data model overview

The core data model is centered around a trip.

```text
users
  |
trip_members
  |
trips
  |--- trip_routes
  |       |
  |       └── locations
  |
  |--- expenses
  |
  └── messages
```

Important entities:

- `users` — registered users;
- `trips` — shared trip workspaces;
- `trip_members` — membership and roles;
- `trip_routes` — route containers inside a trip;
- `locations` — map points with ordering;
- `expenses` — shared expenses;
- `messages` — trip chat messages.

## Security model

The project includes several application-level security mechanisms:

- password hashing with bcrypt;
- JWT-based authentication;
- role checks for trip operations;
- server-side membership validation;
- protected WebSocket connection;
- CORS and Origin validation;
- rate limiting for authentication endpoints;
- request body size limits.

Authorization is based on trip membership. JWT identifies the user, while access to trip data is checked separately through `trip_members` and the user's role inside a specific trip.

## Testing and CI

The project includes automated checks for key backend logic and frontend build correctness.

Covered areas include:

- backend configuration validation;
- authentication-related helpers;
- request limiting behavior;
- WebSocket origin checks;
- expense calculation helpers;
- frontend TypeScript and production build.

GitHub Actions are used to run backend tests and frontend build checks on repository updates.

## Android build

The frontend can also be packaged as an Android application through Capacitor. The Android build uses the same backend API and WebSocket endpoints as the web version, so users, trips, messages and expenses remain shared between web and mobile clients.

## Current status

Implemented and tested:

- authentication;
- trip management;
- invite-based joining;
- roles inside trips;
- route and map point management;
- shared expenses;
- balance calculation;
- trip chat;
- public web version;
- Android build.

## Roadmap

Possible future improvements:

- refresh token flow;
- OAuth login;
- richer notification system;
- export of trip expenses;
- improved mobile UI;
- audit log for important trip actions;
- extended test coverage for frontend components.
