import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKey, type AgendaContext } from '../../api/agenda'
import { cancelReservation, getReservation } from './detalle-reserva.api'

export function useDetalleReserva(context: AgendaContext, id: string) {
  const client = useQueryClient()
  const key = queryKey(context, 'detalle', id)
  const query = useQuery({ queryKey: key, queryFn: () => getReservation(context, id) })
  const cancel = useMutation({
    mutationFn: (version: number) => cancelReservation(context, id, version),
    onSuccess: value => client.setQueryData(key, value),
  })
  return { query, cancel }
}
