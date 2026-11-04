import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/user-agreement')({
  component: () => (
    <PlaceholderPage
      title='Terms of Service'
      description='The terms of service will be published before the public launch.'
    />
  ),
})
