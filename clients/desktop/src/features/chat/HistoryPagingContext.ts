import { createContext, type ReactNode } from 'react';

// Feature-owned paging is projected into the work component's presentation slots.
export const HistoryPagingContext = createContext<((direction: 'before' | 'after') => ReactNode) | undefined>(undefined);
