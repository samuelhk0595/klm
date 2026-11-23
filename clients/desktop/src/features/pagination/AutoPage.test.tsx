import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { AutoPage } from './AutoPage';

afterEach(cleanup);

it('does not add a loading row during background work and retains Retry', () => {
  const retry = vi.fn();
  const { rerender } = render(<AutoPage hasMore cursor="page" busy loading={false} error="Offline" onLoad={vi.fn()} onRetry={retry} />);
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
  expect(screen.getByRole('alert')).toHaveTextContent('Offline');
  expect(screen.getByRole('button', { name: 'Retry' })).toBeDisabled();
  rerender(<AutoPage hasMore cursor="page" loading={false} onLoad={vi.fn()} onRetry={retry} />);
  expect(screen.getByRole('alert')).toHaveTextContent('Offline');
  fireEvent.click(screen.getByRole('button', { name: 'Retry' }));
  expect(retry).toHaveBeenCalledOnce();
});

it('still shows progress for ordinary pagination by default', () => {
  render(<AutoPage hasMore cursor="page" busy onLoad={vi.fn()} />);
  expect(screen.getByRole('status')).toHaveTextContent('Loading…');
});
