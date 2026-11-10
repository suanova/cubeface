import { createFileRoute } from '@tanstack/react-router'

import { ModelLibraryPage } from '@/features/models/components/model-library-page'
import { modelSearchSchema } from '@/features/models/search'

export const Route = createFileRoute('/models/')({
  validateSearch: modelSearchSchema,
  component: ModelLibraryPage,
})
