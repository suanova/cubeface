import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/rankings')({
  component: () => (
    <PlaceholderPage
      title='Rankings'
      description='Model leaderboards and usage rankings are coming soon.'
    />
  ),
})
