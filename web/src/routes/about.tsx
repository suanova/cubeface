import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/about')({
  component: () => (
    <PlaceholderPage
      title='About CubeFace'
      description='Learn more about the project, its mission, and the people behind it.'
    />
  ),
})
