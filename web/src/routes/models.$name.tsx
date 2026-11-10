import { createFileRoute } from '@tanstack/react-router'

import { ModelDetailPage } from '@/features/models/components/model-detail-page'

export const Route = createFileRoute('/models/$name')({
  component: ModelDetailPage,
})
