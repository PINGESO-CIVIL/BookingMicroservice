import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { queryKey, type AgendaContext } from '../../api/agenda'
import { reservationList, resourceNames, type ReservationFilters } from './lista-reservas.api'

export function useListaReservas(context: AgendaContext, admin: boolean, filters: ReservationFilters) {
  return {
    reservations: useInfiniteQuery({
      queryKey: queryKey(context, 'reservas', admin, filters),
      initialPageParam: undefined as string | undefined,
      queryFn: ({ pageParam }) => reservationList(context, admin, { ...filters, cursor: pageParam }),
      getNextPageParam: page => page.nextCursor ?? undefined,
    }),
    resources: useQuery({ queryKey: queryKey(context, 'nombres'), queryFn: () => resourceNames(context) }),
  }
}
