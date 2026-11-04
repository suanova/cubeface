import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/models')({
  component: () => (
    <PlaceholderPage
      title='Model hub'
      description='Browse and publish models. This catalog is being built.'
    />
  ),
})
