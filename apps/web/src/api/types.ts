import type { components } from './schema'
export type Schemas = components['schemas']
export type Reservation = Schemas['Reservation']
export type Resource = Schemas['Resource']
export type ResourceDetail = Schemas['ResourceDetail']
export type Policy = Schemas['ResourcePolicy']
export type ScheduleRule = Schemas['WeeklyScheduleRule']
export type Closure = Schemas['ResourceClosure']
export type Availability = Schemas['Availability']
export type Page<T> = { items: T[]; nextCursor: string | null }
