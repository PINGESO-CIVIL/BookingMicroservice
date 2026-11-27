import { useInfiniteQuery } from '@tanstack/react-query'
import { queryKey, type AgendaContext } from '../../api/agenda'
import { resourceList } from './lista-recursos.api'

export function useListaRecursos(context: AgendaContext, enabled: boolean) {
  return useInfiniteQuery({
    queryKey: queryKey(context, 'recursos', enabled), initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) => resourceList(context, enabled, pageParam),
    getNextPageParam: page => page.nextCursor ?? undefined,
  })
}
