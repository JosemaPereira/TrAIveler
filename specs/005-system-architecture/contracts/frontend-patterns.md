# Frontend Architectural Patterns

**Feature**: System Architecture and Technology Stack  
**Created**: 2026-07-03  
**Layer**: Frontend (React SPA)

## Overview

This document defines architectural patterns for frontend development: component layering (Atomic Design), data fetching (TanStack Query), state management (Zustand), styling (CSS Modules + design tokens), and API integration (fetch wrapper). All patterns enforce WCAG 2.1 AA compliance and testability.

---

## Pattern 1: Atomic Design Component Layering

### Structure

```
src/components/
├── primitives/        # Atoms: Button, Input, Card, Modal
├── composites/        # Molecules: Form, DataTable, NavigationBar
└── (no organisms here — those belong in features/)

src/features/<feature-name>/
└── components/        # Organisms & Templates: feature-specific compositions
```

### Primitives (Atoms)

**Characteristics**:
- Single responsibility (e.g., Button renders a clickable button)
- No business logic or API calls
- Configurable via props only
- Reusable across all features
- Must use design tokens for styling
- Must handle accessibility (ARIA, keyboard, focus)

**Example** (`src/components/primitives/Button.tsx`):
```tsx
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  onClick?: () => void;
  children: React.ReactNode;
  'aria-label'?: string;
}

export function Button({ variant, size = 'md', disabled, onClick, children, ...aria }: ButtonProps) {
  return (
    <button
      className={styles[`button-${variant}-${size}`]}
      disabled={disabled}
      onClick={onClick}
      {...aria}
    >
      {children}
    </button>
  );
}
```

**Styling** (`Button.module.css`):
```css
.button-primary-md {
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
  padding: var(--space-md) var(--space-lg);
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  /* NO hard-coded values */
}
```

---

### Composites (Molecules)

**Characteristics**:
- Compose multiple primitives
- Handle simple interactions (e.g., form validation)
- May use local state (`useState`)
- Reusable across multiple features
- Must handle Loading, Error, Empty states if data-dependent

**Example** (`src/components/composites/Form.tsx`):
```tsx
interface FormProps {
  onSubmit: (data: Record<string, unknown>) => Promise<void>;
  children: React.ReactNode;
  isLoading?: boolean;
}

export function Form({ onSubmit, children, isLoading }: FormProps) {
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data = Object.fromEntries(formData);

    try {
      await onSubmit(data);
    } catch (err) {
      setErrors({ _form: err.message });
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {errors._form && <ErrorMessage>{errors._form}</ErrorMessage>}
      {children}
      <Button type="submit" variant="primary" disabled={isLoading}>
        {isLoading ? 'Submitting...' : 'Submit'}
      </Button>
    </form>
  );
}
```

---

### Features (Organisms & Templates)

**Characteristics**:
- Feature-specific business logic
- Integrate with API via TanStack Query hooks
- Use global state via Zustand stores
- Compose primitives and composites
- Handle full Loading, Error, Empty states

**Example** (`src/features/trips/components/TripList.tsx`):
```tsx
export function TripList() {
  const { data: trips, isLoading, error } = useQuery({
    queryKey: ['trips'],
    queryFn: () => apiRequest<Trip[]>('/trips'),
  });

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage>Failed to load trips</ErrorMessage>;
  if (!trips || trips.length === 0) return <EmptyState>No trips found</EmptyState>;

  return (
    <div className={styles.tripList}>
      {trips.map(trip => (
        <TripCard key={trip.id} trip={trip} />
      ))}
    </div>
  );
}
```

---

## Pattern 2: TanStack Query for API Communication

### Query Key Convention

**Structure**: `[resource, ...identifiers, ...filters]`

**Examples**:
```tsx
['trips']                           // List all trips
['trips', tripId]                   // Single trip by ID
['trips', tripId, 'activities']     // Activities for a trip
['suggestions', { status: 'pending' }] // Suggestions filtered by status
```

### Query Hook Pattern

**Location**: `src/features/<feature>/hooks/use<Resource>Query.ts`

**Example** (`src/features/trips/hooks/useTripQuery.ts`):
```tsx
export function useTripQuery(tripId: string) {
  return useQuery({
    queryKey: ['trips', tripId],
    queryFn: () => apiRequest<Trip>(`/trips/${tripId}`),
    staleTime: 1000 * 60 * 5, // 5 minutes
    enabled: !!tripId,         // Only run if tripId provided
  });
}
```

### Mutation Hook Pattern

**Example** (`src/features/trips/hooks/useCreateTripMutation.ts`):
```tsx
export function useCreateTripMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTripRequest) =>
      apiRequest<Trip>('/trips', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      // Invalidate trips list to refetch
      queryClient.invalidateQueries({ queryKey: ['trips'] });
    },
  });
}
```

### Error Handling

**Pattern**: TanStack Query catches errors thrown by `apiRequest`. Components handle via `error` prop.

```tsx
export function TripDetail({ tripId }: { tripId: string }) {
  const { data: trip, error, refetch } = useTripQuery(tripId);

  if (error) {
    return (
      <ErrorMessage>
        {error.message}
        <Button onClick={() => refetch()}>Retry</Button>
      </ErrorMessage>
    );
  }

  // ... render trip
}
```

---

## Pattern 3: Zustand for Global State

### Store Definition

**Location**: `src/stores/<domain>-store.ts`

**Example** (`src/stores/auth-store.ts`):
```tsx
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
  refreshSession: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,

  login: (user) => set({ isAuthenticated: true, user }),

  logout: () => {
    set({ isAuthenticated: false, user: null });
    // Clear session storage
    sessionStorage.removeItem('auth');
  },

  refreshSession: async () => {
    try {
      const user = await apiRequest<User>('/auth/me');
      set({ isAuthenticated: true, user });
    } catch (err) {
      set({ isAuthenticated: false, user: null });
    }
  },
}));
```

### Store Usage

**Pattern**: Use typed selectors to avoid unnecessary re-renders.

```tsx
// ❌ Bad: Component re-renders on any auth state change
const authState = useAuthStore();

// ✅ Good: Component only re-renders when isAuthenticated changes
const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
```

**Persistence** (if needed):
```tsx
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({ /* state and actions */ }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => sessionStorage), // NOT localStorage (security)
    }
  )
);
```

---

## Pattern 4: API Client Integration

### Client Configuration

**Location**: `src/lib/api-client.ts`

**Implementation**:
```tsx
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public requestId?: string
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export async function apiRequest<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Request-ID': crypto.randomUUID(),
      ...options?.headers,
    },
    credentials: 'include', // Send HTTP-only cookies (auth tokens)
  });

  if (!response.ok) {
    const error = await response.json();
    throw new APIError(
      error.message || 'Request failed',
      response.status,
      error.request_id
    );
  }

  return response.json();
}
```

### Error Handling in Components

**Pattern**: Handle errors at component level, display user-friendly messages.

```tsx
export function TripList() {
  const { data, error } = useQuery({
    queryKey: ['trips'],
    queryFn: () => apiRequest<Trip[]>('/trips'),
  });

  if (error) {
    const message =
      error.status === 401
        ? 'Please log in to view trips'
        : error.status === 403
        ? 'You do not have permission to view trips'
        : 'Failed to load trips. Please try again.';

    return <ErrorMessage>{message}</ErrorMessage>;
  }

  // ... render data
}
```

---

## Pattern 5: Design Tokens and CSS Modules

### Token Usage

**Rule**: All styling MUST reference CSS custom properties from `src/styles/tokens.css`. Hard-coded values forbidden.

**Component CSS** (`TripCard.module.css`):
```css
.tripCard {
  background-color: var(--color-surface);
  padding: var(--space-lg);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  color: var(--color-text);
}

.tripCard:hover {
  box-shadow: var(--shadow-md);
}

.tripTitle {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  margin-bottom: var(--space-sm);
}
```

### Component Implementation

```tsx
import styles from './TripCard.module.css';

export function TripCard({ trip }: { trip: Trip }) {
  return (
    <div className={styles.tripCard}>
      <h3 className={styles.tripTitle}>{trip.title}</h3>
      <p>{trip.description}</p>
    </div>
  );
}
```

### Conditional Styling

**Pattern**: Use template literals for dynamic classes, not inline styles.

```tsx
<button className={`${styles.button} ${variant === 'primary' ? styles.buttonPrimary : styles.buttonSecondary}`}>
  {children}
</button>
```

---

## Pattern 6: Accessibility Requirements

### Keyboard Operability

**Rule**: All interactive elements MUST be keyboard-accessible.

```tsx
export function TripCard({ trip, onDelete }: TripCardProps) {
  return (
    <div className={styles.tripCard}>
      <button
        onClick={onDelete}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onDelete();
          }
        }}
        aria-label={`Delete trip ${trip.title}`}
      >
        Delete
      </button>
    </div>
  );
}
```

### Focus Indicators

**Rule**: All focusable elements MUST have visible focus indicators.

```css
.button:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

### ARIA Labels

**Rule**: All form inputs MUST have associated labels. Buttons with icons MUST have `aria-label`.

```tsx
<label htmlFor="trip-title">Trip Title</label>
<Input id="trip-title" name="title" required />

<Button aria-label="Close dialog" onClick={onClose}>
  <CloseIcon />
</Button>
```

---

## Pattern 7: Loading, Error, Empty States

### Mandatory States

Every data-dependent component MUST handle:
1. **Loading**: Show spinner or skeleton
2. **Error**: Show error message with retry action
3. **Empty**: Show "no data" message with call-to-action

**Example**:
```tsx
export function TripList() {
  const { data: trips, isLoading, error, refetch } = useQuery({
    queryKey: ['trips'],
    queryFn: () => apiRequest<Trip[]>('/trips'),
  });

  if (isLoading) {
    return <LoadingSpinner aria-label="Loading trips" />;
  }

  if (error) {
    return (
      <ErrorMessage>
        Failed to load trips.
        <Button onClick={() => refetch()}>Retry</Button>
      </ErrorMessage>
    );
  }

  if (!trips || trips.length === 0) {
    return (
      <EmptyState>
        No trips found.
        <Button onClick={() => navigate('/trips/new')}>Create Trip</Button>
      </EmptyState>
    );
  }

  return (
    <div className={styles.tripList}>
      {trips.map(trip => <TripCard key={trip.id} trip={trip} />)}
    </div>
  );
}
```

---

## Pattern 8: Testing Patterns

### Component Testing with React Testing Library

**Location**: `src/features/<feature>/components/__tests__/Component.test.tsx`

**Example**:
```tsx
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TripList } from '../TripList';
import { server } from '@/tests/__mocks__/server';
import { rest } from 'msw';

describe('TripList', () => {
  it('renders loading state initially', () => {
    render(<TripList />);
    expect(screen.getByLabelText('Loading trips')).toBeInTheDocument();
  });

  it('renders trips after loading', async () => {
    server.use(
      rest.get('/trips', (req, res, ctx) => {
        return res(ctx.json([
          { id: '1', title: 'Paris Trip' },
          { id: '2', title: 'Tokyo Trip' },
        ]));
      })
    );

    render(<TripList />);

    await waitFor(() => {
      expect(screen.getByText('Paris Trip')).toBeInTheDocument();
      expect(screen.getByText('Tokyo Trip')).toBeInTheDocument();
    });
  });

  it('renders error state on API failure', async () => {
    server.use(
      rest.get('/trips', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({ message: 'Server error' }));
      })
    );

    render(<TripList />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load trips/i)).toBeInTheDocument();
    });
  });
});
```

### MSW Mock Setup

**Location**: `src/tests/__mocks__/server.ts`

```tsx
import { setupServer } from 'msw/node';
import { rest } from 'msw';

export const handlers = [
  rest.get('/trips', (req, res, ctx) => {
    return res(ctx.json([]));
  }),
];

export const server = setupServer(...handlers);
```

---

## Summary

Frontend architectural patterns enforce:
- **Component layering**: Primitives → Composites → Features (Atomic Design)
- **Data fetching**: TanStack Query with typed hooks and query key conventions
- **Global state**: Zustand stores with typed selectors (no prop drilling)
- **Styling**: CSS Modules + design tokens (no hard-coded values)
- **Accessibility**: WCAG 2.1 AA compliance (keyboard, ARIA, focus indicators)
- **Error handling**: Loading, Error, Empty states mandatory for all data components
- **Testability**: React Testing Library + MSW for component tests

All patterns align with constitution principles: accessible & token-driven UI (principle IV), code quality (principle III), simplicity (principle II).
