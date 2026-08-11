import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import Page from '../routes/+page.svelte';

describe('root page', () => {
  it('shows the frontend wired status', () => {
    render(Page);

    expect(screen.getByText('frontend wired')).toBeInTheDocument();
  });
});
