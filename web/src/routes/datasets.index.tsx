import { createFileRoute } from '@tanstack/react-router'

import { DatasetLibraryPage } from '@/features/datasets/components/dataset-library-page'
import { datasetSearchSchema } from '@/features/datasets/search'

export const Route = createFileRoute('/datasets/')({
  validateSearch: datasetSearchSchema,
  component: DatasetLibraryPage,
})
