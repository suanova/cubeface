import { createFileRoute } from '@tanstack/react-router'

import { PlaceholderPage } from '@/components/placeholder-page'

export const Route = createFileRoute('/privacy-policy')({
  component: () => (
    <PlaceholderPage
      title='Privacy Policy'
      description='The privacy policy will be published before the public launch.'
    />
  ),
})
