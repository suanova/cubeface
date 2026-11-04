import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/pricing')({
  component: () => (
    <PlaceholderPage
      title='Pricing'
      description='Plans and usage-based pricing will be announced before launch.'
    />
  ),
})
