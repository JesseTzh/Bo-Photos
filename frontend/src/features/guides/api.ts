import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";

export interface Guide { id:string;title:string;country:string;city:string;days:number;start_date?:string;end_date?:string;cover_asset_id?:string;cover_url?:string;published:boolean;sort:number }
export interface GuideModule { id:string;guide_id:string;name:string;kind:"content"|"structured";template?:string;data_version:number;structured_data?:unknown;sort:number;hidden:boolean;blocks?:GuideBlock[] }
export interface GuideBlock { id:string;module_id:string;type:"markdown"|"image"|"video"|"link"|"tasks"|"warning"|"divider";data_version:number;data:unknown;sort:number }
export interface TOCItem { id:string;title:string;level:1|2;target_module_id:string;sort:number;hidden:boolean }
export interface GuideAlbum { id:string;name:string;value:string;coverURL:string;sort:number }
export interface GuideDetail { guide:Guide;modules:GuideModule[];toc:TOCItem[];albums:GuideAlbum[] }
export type GuideInput=Omit<Guide,"id"|"cover_url">;
export function useGuides(admin=false){return useQuery({queryKey:["guides",admin],queryFn:()=>apiRequest<{items:Guide[]}>(`${admin?"/admin":"/public"}/guides`)})}
export function useGuide(id?:string,admin=false){return useQuery({queryKey:["guide",id,admin],queryFn:()=>apiRequest<GuideDetail>(`${admin?"/admin":"/public"}/guides/${id}`),enabled:Boolean(id)})}
function useGuideMutation<T>(fn:(input:T)=>Promise<unknown>){const client=useQueryClient();return useMutation({mutationFn:fn,onSuccess:()=>client.invalidateQueries({queryKey:["guides"]})})}
export function useSaveGuide(){return useGuideMutation(({id,input}:{id?:string;input:GuideInput})=>apiRequest<Guide>(id?`/admin/guides/${id}`:"/admin/guides",{method:id?"PUT":"POST",body:JSON.stringify(input)}))}
export function useDeleteGuide(){return useGuideMutation((id:string)=>apiRequest<void>(`/admin/guides/${id}`,{method:"DELETE"}))}
export function useCreateModule(){return useGuideMutation(({guideId,input}:{guideId:string;input:Record<string,unknown>})=>apiRequest(`/admin/guides/${guideId}/modules`,{method:"POST",body:JSON.stringify(input)}))}
export function useCreateBlock(){return useGuideMutation(({guideId,moduleId,input}:{guideId:string;moduleId:string;input:Record<string,unknown>})=>apiRequest(`/admin/guides/${guideId}/modules/${moduleId}/blocks`,{method:"POST",body:JSON.stringify(input)}))}
export function useAutoTOC(){return useGuideMutation((guideId:string)=>apiRequest<void>(`/admin/guides/${guideId}/toc/auto-generate`,{method:"POST"}))}
export function useUpdateModule(){return useGuideMutation(({guideId,moduleId,input}:{guideId:string;moduleId:string;input:Record<string,unknown>})=>apiRequest<void>(`/admin/guides/${guideId}/modules/${moduleId}`,{method:"PUT",body:JSON.stringify(input)}))}
export function useDeleteModule(){return useGuideMutation(({guideId,moduleId}:{guideId:string;moduleId:string})=>apiRequest<void>(`/admin/guides/${guideId}/modules/${moduleId}`,{method:"DELETE"}))}
export function useUpdateBlock(){return useGuideMutation(({guideId,moduleId,blockId,input}:{guideId:string;moduleId:string;blockId:string;input:Record<string,unknown>})=>apiRequest<void>(`/admin/guides/${guideId}/modules/${moduleId}/blocks/${blockId}`,{method:"PUT",body:JSON.stringify(input)}))}
export function useReplaceTOC(){return useGuideMutation(({guideId,items}:{guideId:string;items:TOCItem[]})=>apiRequest<void>(`/admin/guides/${guideId}/toc`,{method:"PUT",body:JSON.stringify({items})}))}
export function useReplaceGuideAlbums(){return useGuideMutation(({guideId,ids}:{guideId:string;ids:string[]})=>apiRequest<void>(`/admin/guides/${guideId}/albums`,{method:"PUT",body:JSON.stringify({ids})}))}
export function useDeleteBlock(){return useGuideMutation(({guideId,moduleId,blockId}:{guideId:string;moduleId:string;blockId:string})=>apiRequest<void>(`/admin/guides/${guideId}/modules/${moduleId}/blocks/${blockId}`,{method:"DELETE"}))}
export function useSortModules(){return useGuideMutation(({guideId,ids}:{guideId:string;ids:string[]})=>apiRequest<void>(`/admin/guides/${guideId}/modules/sort`,{method:"PUT",body:JSON.stringify({ids})}))}
export function useSortBlocks(){return useGuideMutation(({guideId,moduleId,ids}:{guideId:string;moduleId:string;ids:string[]})=>apiRequest<void>(`/admin/guides/${guideId}/modules/${moduleId}/blocks/sort`,{method:"PUT",body:JSON.stringify({ids})}))}
export function useSortGuides(){return useGuideMutation((ids:string[])=>apiRequest<void>("/admin/guides/sort",{method:"PUT",body:JSON.stringify({ids})}))}
