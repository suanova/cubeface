import { createFileRoute } from '@tanstack/react-router'

import '@/features/models/models.css'
import '@/features/studio/studio.css'
import { StudioPage } from '@/features/studio/studio-page'

export const Route = createFileRoute('/studio')({
  component: StudioPage,
})
