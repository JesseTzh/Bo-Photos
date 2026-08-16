import { useMutation,useQuery,useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";
import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import type { Asset } from "../assets/schema";
export interface SiteSettings{site_title:string;site_author:string;site_favicon_url:string;hero_asset_id:string;hero_carousel_asset_ids:string[];hero_show_text:boolean;hero_eyebrow:string;hero_kicker:string;hero_title:string;hero_accent_title:string;hero_description:string;hero_video_overlay:boolean;about_intro:string;about_social_instagram:string;about_social_xiaohongshu:string;about_social_weibo:string;about_social_github:string;about_gallery_asset_ids:string[];gallery_layout:"grid"|"single";public_original_download:boolean;admin_images_per_page:number;max_upload_files:number;preview_quality:number;preview_max_width:number;analytics_enabled:boolean;analytics_retention_days:number;analytics_timezone:string}
export interface CountPoint{date:string;count:number}
export interface NamedCount{name:string;count:number}
export interface Dashboard{ImagesTotal:number;ImagesPublic:number;AlbumsTotal:number;VisitsTotal:number;VisitsToday:number;VisitsYesterday:number;CamerasTotal:number;LensesTotal:number;last_7_days:CountPoint[];TopCameras:NamedCount[];TopLenses:NamedCount[];PhotosByYear:NamedCount[]}
export interface Analytics{dashboard:Dashboard;unique_visitors:number;hourly:CountPoint[];sources:NamedCount[];pages:NamedCount[]}
export function usePublicSettings(){return useQuery({queryKey:["settings","public"],queryFn:()=>apiRequest<Partial<SiteSettings>>("/public/settings")})}
export function useAdminSettings(){return useQuery({queryKey:["settings","admin"],queryFn:()=>apiRequest<SiteSettings>("/admin/settings")})}
export function useSaveSettings(){const c=useQueryClient();return useMutation({mutationFn:(s:SiteSettings)=>apiRequest<void>("/admin/settings",{method:"PUT",body:JSON.stringify(s)}),onSuccess:()=>c.invalidateQueries({queryKey:["settings"]})})}
export function useHomeAssets(ids:string[]){return useQuery({queryKey:["assets","home",ids],queryFn:async()=>{const items=await Promise.all(ids.map((id)=>apiRequest<Asset>(`/public/assets/${id}`).catch(()=>undefined)));return items.filter((item):item is Asset=>Boolean(item))},enabled:ids.length>0})}
export function useDashboard(){return useQuery({queryKey:["dashboard"],queryFn:()=>apiRequest<Dashboard>("/admin/dashboard")})}
export function useAnalytics(){return useQuery({queryKey:["analytics"],queryFn:()=>apiRequest<Analytics>("/admin/analytics")})}
export function useDisk(){return useQuery({queryKey:["disk"],queryFn:()=>apiRequest<{total:number;free:number;used:number}>("/admin/disk")})}
export function recordVisit(path:string,pageType:string){return apiRequest<void>("/public/visits",{method:"POST",body:JSON.stringify({path,page_type:pageType,referrer:document.referrer})})}
export function useVisit(pageType:string){const location=useLocation();useEffect(()=>{void recordVisit(location.pathname,pageType).catch(()=>undefined)},[location.pathname,pageType])}
