import { Toaster as Sonner, type ToasterProps } from 'sonner'

export function Toaster(props: ToasterProps) {
  return <Sonner position='top-center' richColors closeButton {...props} />
}
