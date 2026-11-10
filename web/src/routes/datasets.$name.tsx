import { createFileRoute } from '@tanstack/react-router'

import { DatasetDetailPage } from '@/features/datasets/components/dataset-detail-page'

export const Route = createFileRoute('/datasets/$name')({
  component: DatasetDetailPage,
})
