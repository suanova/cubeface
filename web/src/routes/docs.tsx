import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/docs')({
  component: () => (
    <PlaceholderPage
      title='Documentation'
      description='Guides and API reference are on the way.'
    />
  ),
})
