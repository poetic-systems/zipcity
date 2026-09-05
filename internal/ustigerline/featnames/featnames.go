package featnames

import (
	"maps"
	"strings"

	"github.com/poetic-systems/zipcity/internal/ustigerline/directionals"
	"github.com/poetic-systems/zipcity/internal/ustigerline/fieldutil"
	"github.com/poetic-systems/zipcity/internal/ustigerline/qualifiers"
)

type FeatnameInfo struct {
	Code        string
	Full        string
	Short       string
	Spanish     bool
	Translation string
	Prefix      bool
	Suffix      bool
}

/*
TypeCode ExpandedFullText DisplayNameAbbreviation Spanish SpanishTranslation PrefixType SuffixType
103 Academy Acdmy No N/A Y Y
104 Acueducto Acueducto Yes Aqueduct Y N
105 Aeropuerto Aero Yes Airport Y N
106 Air Force Base AFB No N/A N Y
107 Airfield Airfield No N/A N Y
108 Airpark Airpark No N/A N Y
109 Airport Arprt No N/A N Y
110 Airstrip Airstrip No N/A N Y
112 Alley Aly No N/A N Y
115 Apartment Building Apt Bldg No N/A N Y
116 Apartment Complex Apt Complex No N/A N Y
117 Apartments Apts No N/A N Y
118 Aqueduct Aqueduct No N/A N Y
119 Arcade Arc No N/A Y Y
121 Arroyo Arroyo Yes Stream Y N
122 Assisted Living Center Asstd Liv Ctr No N/A N Y
694 Assisted Living Facility Asstd Liv Fac No N/A N Y
123 Autopista Autopista Yes Expressway or Freeway Y N
124 Avenida Ave Yes Avenue Y N
125 Avenue Ave No N/A Y Y
126 Bahia Bahía Yes Bay Y N
127 Bank Bk No N/A Y Y
704 Base Base No N/A N Y
128 Basin Basin No N/A N Y
129 Bay Bay No N/A Y Y
130 Bayou Byu No N/A Y Y
131 Beach Bch No N/A N Y
132 Bed and Breakfast B and B No N/A N Y
136 Beltway Beltway No N/A N Y
137 Bend Bnd No N/A N Y
138 Bluff Blf No N/A N Y
139 Boarding House Brdng Hse No N/A N Y
140 Bog Bog No N/A N Y
141 Bosque Bosque Yes Forest Y N
142 Boulevard Blvd No N/A Y Y
143 Boundary Boundary No N/A N Y
146 Branch Br No N/A Y Y
147 Bridge Brg No N/A N Y
148 Brook Brk No N/A N Y
149 Building Bldg No N/A Y Y
150 Bulevar Bulevar Yes Boulevard Y N
151 Bureau of Indian Affairs Highway BIA Hwy No N/A Y N
152 Bureau of Indian Affairs Road BIA Rd No N/A Y N
153 Bureau of Indian Affairs Route BIA Rte No N/A Y N
154 Bureau of Land Management Road BLM Rd No N/A Y N
696 Bypass Byp No N/A Y Y
156 Calle Cll Yes Street Y N
157 Calleja Calleja Yes Narrow Street Y N
158 Callejón Callejón Yes Alley Y N
159 Caminito Cmt Yes Little Road Y N
160 Camino Cam Yes Road or Way Y N
161 Camp Cp No N/A Y Y
163 Campground Cmpgrnd No N/A N Y
164 Campus Cmps No N/A N Y
165 Canal Cnl No N/A Y Y
172 Cano Caño Yes Drain or Sewer Y N
166 Cantera Cantera Yes Quarry or Gravel Pit Y N
167 Canyon Cyn No N/A Y Y
168 Capilla Capilla Yes Chapel Y N
169 Carretera Carr Yes Road Y N
170 Causeway Cswy No N/A N Y
171 Cayo Cayo Yes Key Y N
173 Cementerio Cem Yes Cemetery Y N
174 Cemetery Cmtry No N/A N Y
175 Center Ctr No N/A Y Y
176 Centro Centro Yes Center Y N
177 Cerrada Cer Yes Closed Y N
178 Chamber of Commerce Cham of Com No N/A N Y
179 Channel Chnnl No N/A N Y
180 Chapel Cpl No N/A Y Y
181 Childrens Home Childrens Home No N/A N Y
182 Church Church No N/A Y Y
183 Circle Cir No N/A N Y
234 Círculo Cír Yes Circle Y N
184 City Hall City Hall No N/A N Y
185 City Park City Park No N/A N Y
186 Cliff Clf No N/A N Y
187 Club Clb No N/A Y Y
188 Colegio Colegio Yes School Y N
189 College Colg No N/A Y Y
190 Common Cmn No N/A N Y
191 Commons Cmns No N/A Y Y
192 Community Center Community Ctr No N/A N Y
193 Community College Community Colg No N/A Y Y
194 Community Park Community Park No N/A Y Y
195 Complex Complx No N/A N Y
197 Condominios Condios Yes Condominiums Y N
198 Condominium Condo No N/A Y Y
199 Condominiums Condos No N/A N Y
201 Convent Cnvnt No N/A Y Y
202 Convention Center Convention Ctr No N/A Y Y
203 Corners Cors No N/A N Y
204 Correctional Facility Corr Faclty No N/A N Y
205 Correctional Institute Corr Inst No N/A N Y
207 Corte Corte Yes Court Y N
679 Cottage Cottage No N/A N Y
208 Coulee Coulee No N/A N Y
209 Country Club Country Club No N/A Y Y
210 County Highway Co Hwy No N/A Y N
211 County Home Co Home No N/A Y Y
212 County Lane Co Ln No N/A Y N
213 County Park Co Park No N/A N Y
214 County Road Co Rd No N/A Y N
215 County Route Co Rte No N/A Y N
216 County State Aid Highway Co St Aid Hwy No N/A Y N
217 County Trunk Highway Co Trunk Hwy No N/A Y N
218 County Trunk Road Co Trunk Rd No N/A Y N
219 Course Crs No N/A N Y
220 Court Ct No N/A Y Y
221 Courthouse Courthouse No N/A N Y
222 Courts Cts No N/A N Y
223 Cove Cv No N/A N Y
225 Creek Crk No N/A N Y
226 Crescent Cres No N/A N Y
227 Crest Crst No N/A N Y
228 Crossing Xing No N/A N Y
229 Crossroads Xroad No N/A Y Y
233 Cutoff Cutoff No N/A N Y
235 Dam Dm No N/A N Y
236 Delta Road Delta Rd No N/A Y N
237 Department Dept No N/A Y Y
238 Depot Dep No N/A N Y
239 Detention Center Detention Ctr No N/A N Y
240 District of Columbia Highway DC Hwy No N/A Y N
241 Ditch Ditch No N/A Y Y
242 Divide Dv No N/A N Y
243 Dock Dock No N/A N Y
244 Dormitory Dormitory No N/A N Y
245 Drain Drn No N/A N Y
246 Draw Draw No N/A N Y
247 Drive Dr No N/A N Y
248 Driveway Driveway No N/A Y Y
249 Dump Dump No N/A N Y
251 Edificio Edif Yes Building Y N
252 Elementary School Elem School No N/A N Y
253 Ensenada Ensenada Yes Cove Y N
254 Entrada Ent Yes Entrance Y N
256 Escuela Escuela Yes School Y N
680 Esplanade Esplanade Yes Esplanade Y Y
257 Estates Ests No N/A N Y
260 Estuary Estuary No N/A N Y
261 Expreso Expreso Yes Expressway Y N
262 Expressway Expy No N/A Y Y
263 Extension Ext No N/A Y Y
264 Facility Faclty No N/A N Y
265 Fairgrounds Fairgrounds No N/A N Y
266 Falls Fls No N/A Y Y
267 Farm Frm No N/A N Y
268 Farm Road Farm Rd No N/A Y N
269 Farm-to-Market Road FM No N/A Y N
275 Fence Line Fence Line No N/A N Y
276 Ferry Crossing Ferry Crossing No N/A Y Y
277 Field Fld No N/A N Y
278 Fire Control Road Fire Cntrl Rd No N/A Y N
279 Fire Department Fire Dept No N/A N Y
280 Fire District Road Fire Dist Rd No N/A Y N
281 Fire Lane Fire Ln No N/A Y N
282 Fire Road Fire Rd No N/A Y N
283 Fire Route Fire Rte No N/A Y N
284 Fire Station Fire Sta No N/A Y Y
285 Fire Trail Fire Trl No N/A Y N
286 Flowage Flowage No N/A N Y
287 Flume Flume No N/A N Y
288 Forest Frst No N/A N Y
289 Forest Highway Forest Hwy No N/A Y Y
290 Forest Road Forest Rd No N/A Y N
291 Forest Route Forest Rte No N/A Y N
292 Forest Service Road FS Rd No N/A Y N
293 Fork Frk No N/A N Y
294 Fort Ft No N/A Y N
295 Four-Wheel Drive Trail 4WD Trl No N/A Y Y
296 Fraternity Frtrnty No N/A N Y
297 Freeway Fwy No N/A N Y
298 Garage Grge No N/A N Y
299 Gardens Gdns No N/A N Y
303 Glacier Glacier No N/A N Y
304 Glen Gln No N/A N Y
305 Golf Club Golf Club No N/A Y Y
306 Golf Course Golf Course No N/A Y Y
307 Grade Grade No N/A N Y
309 Green Grn No N/A N Y
310 Group Home Group Home No N/A N Y
311 Gulch Gulch No N/A N Y
312 Gulf Gulf No N/A Y Y
313 Gully Gully No N/A N Y
314 Halfway House Halfway House No N/A N Y
315 Hall Hall No N/A N Y
316 Harbor Hbr No N/A N Y
317 Heights Hts No N/A N Y
321 High School High School No N/A N Y
322 Highway Hwy No N/A Y Y
323 Hill Hl No N/A N Y
324 Hollow Holw No N/A N Y
325 Home Home No N/A Y Y
326 Hospital Hosp No N/A Y Y
327 Hostel Hostel No N/A N Y
328 Hotel Hotel No N/A Y Y
329 House Hse No N/A Y Y
330 Housing Hsng No N/A Y Y
332 Iglesia Iglesia Yes Church Y N
333 Indian Route Indian Rte No N/A Y N
334 Indian Service Route Indian Svc Rte No N/A Y N
336 Industrial Park Indl Park No N/A N Y
337 Inlet Inlt No N/A N Y
338 Inn Inn No N/A Y Y
339 Institute Inst No N/A Y Y
340 Institution Instn No N/A N Y
341 Instituto Instituto Yes Institute Y N
342 Intermediate School Inter School No N/A N Y
344 Interstate Highway I- No N/A Y N
345 Isla Isla Yes Island Y N
346 Island Is No N/A N Y
347 Islands Iss No N/A Y Y
348 Isle Isle No N/A Y Y
349 Jail Jail No N/A N Y
351 Jeep Trail Jeep Trl No N/A Y Y
352 Junction Junction No N/A N Y
353 Junior High School Jr HS No N/A N Y
356 Kill Kill No N/A Y Y
357 Lago Lago Yes Lake Y N
358 Lagoon Lagoon No N/A N Y
360 Laguna Laguna Yes Lagoon Y N
361 Lake Lk No N/A Y Y
362 Lakes Lks No N/A N Y
363 Landfill Lndfll No N/A N Y
364 Landing Lndg No N/A N Y
365 Landing Area Landing Area No N/A Y Y
366 Landing Field Landing Fld No N/A Y Y
367 Landing Strip Landing Strp No N/A Y Y
368 Lane Ln No N/A Y Y
369 Lateral Lateral No N/A Y Y
370 Levee Levee No N/A Y Y
371 Library Lbry No N/A Y Y
372 Lift Lift No N/A Y Y
373 Lighthouse Lighthouse No N/A N Y
374 Line Line No N/A Y Y
376 Lodge Ldg No N/A N Y
377 Logging Road Logging Rd No N/A Y Y
378 Loop Loop No N/A Y Y
379 Mall Mall No N/A Y Y
380 Manor Mnr No N/A N Y
381 Mar Mar Yes Sea Y N
382 Marginal Marginal Yes Service Road Y N
383 Marina Mrna No N/A N Y
384 Marsh Marsh No N/A N Y
385 Meadows Mdws No N/A N Y
386 Medical Building Medical Bldg No N/A N Y
387 Medical Center Medical Ctr No N/A Y Y
388 Memorial Meml No N/A N Y
389 Memorial Gardens Memorial Gnds No N/A N Y
390 Memorial Park Memorial Pk No N/A N Y
391 Mesa Mesa No N/A Y Y
392 Middle School Mid Schl No N/A N Y
393 Military Reservation Mil Res No N/A N Y
394 Millpond Millpond No N/A N Y
395 Mine Mine No N/A N Y
396 Mission Mssn No N/A Y Y
397 Mobile Home Community Mobile Hm Cmty No N/A Y Y
398 Mobile Home Estates Mobile Hm Est No N/A Y Y
399 Mobile Home Park Mobile Hm Pk No N/A Y Y
400 Monastery Monstry No N/A Y Y
401 Monument Mnmt No N/A N Y
403 Mosque Mosque No N/A Y Y
404 Motel Mtl No N/A Y Y
405 Motor Lodge Motor Lodge No N/A N Y
406 Motorway Mtwy No N/A N Y
407 Mount Mt No N/A Y Y
408 Mountain Mtn No N/A N Y
411 Museum Mus No N/A Y Y
412 National Battlefield Natl Bfld No N/A N Y
413 National Battlefield Park Natl Bfld Pk No N/A N Y
414 National Battlefield Site Natl Bfld Site No N/A N Y
415 National Conservation Area Natl Cnsv Area No N/A N Y
416 National Forest Natl Forest No N/A N Y
417 National Forest Development Road Nat For Dev Rd No N/A Y N
419 National Grasslands Natl Grsslnds No N/A N Y
420 National Historic Site Natl Hist Site No N/A N Y
421 National Historical Park Natl Hist Pk No N/A N Y
422 National Lakeshore Natl Lkshr No N/A N Y
423 National Memorial Natl Meml No N/A N Y
424 National Military Park Natl Mil Pk No N/A N Y
425 National Monument Natl Mnmt No N/A N Y
426 National Park Natl Pk No N/A N Y
427 National Preserve Natl Prsv No N/A N Y
428 National Recreation Area Natl Rec Area No N/A N Y
429 National Recreational River Natl Rec Riv No N/A N Y
430 National Reserve Natl Resv No N/A N Y
431 National River Natl Riv No N/A N Y
432 National Scenic Area Natl Sc Area No N/A N Y
433 National Scenic River Natl Sc Riv No N/A N Y
435 National Scenic Riverways Natl Sc Rvrwys No N/A N Y
436 National Scenic Trail Natl Sc Trl No N/A N Y
437 National Seashore Natl Shr No N/A N Y
438 National Wildlife Refuge Natl Wld Rfg No N/A N Y
439 Navajo Service Route Navajo Svc Rte No N/A Y N
440 Naval Air Station Naval Air Sta No N/A N Y
442 Nursing Home Nurse Home No N/A N Y
444 Ocean Ocean No N/A N Y
445 Oceano Océano Yes Ocean Y N
446 Office Ofc No N/A Y Y
447 Office Building Office Bldg No N/A N Y
449 Office Park Office Park No N/A N Y
698 Orchard Orchard No N/A N Y
451 Orchards Orchrds No N/A N Y
452 Orphanage Orphanage No N/A N Y
453 Outlet Outlet No N/A N Y
454 Oval Oval No N/A N Y
455 Overpass Opas No N/A N Y
456 Parish Road Parish Rd No N/A Y N
457 Park Park No N/A N Y
458 Park and Ride Park and Ride No N/A N Y
460 Parkway Pkwy No N/A N Y
706 Parq Parq Yes Park Y N
461 Parque Parque Yes Park Y N
462 Pasaje Pasaje Yes Passage Y N
463 Paseo Pso Yes Path Y N
464 Pass Pass No N/A Y Y
465 Passage Psge No N/A Y Y
466 Path Path No N/A N Y
682 Pavilion Pavilion No N/A N Y
467 Peak Peak No N/A N Y
705 Penitentiary Penitentiary No N/A N Y
468 Pier Pier No N/A Y Y
469 Pike Pike No N/A N Y
470 Pipeline Pipeline No N/A N Y
472 Place Pl No N/A N Y
473 Placita Pla Yes Little Plaza Y N
474 Plant Plnt No N/A N Y
683 Plantation Plantation No N/A N Y
475 Playa Playa Yes Beach Y N
476 Playground Playground No N/A N Y
477 Plaza Plz No N/A Y Y
478 Point Pt No N/A Y Y
479 Pointe Pointe No N/A N Y
480 Police Department Police Dept No N/A Y Y
481 Police Station Police Station No N/A Y Y
482 Pond Pond No N/A Y Y
483 Ponds Ponds No N/A N Y
485 Port Prt No N/A Y Y
486 Post Office Post Office No N/A N Y
487 Power Line Power Line No N/A N Y
691 Power Plant Power Plant No N/A N Y
488 Prairie Pr No N/A N Y
489 Preserve Preserve No N/A N Y
491 Prison Prison No N/A N Y
690 Prison Farm Prison Farm No N/A N Y
685 Promenade Promenade No N/A N Y
492 Prong Prong No N/A N Y
494 Puente Puente Yes Bridge Y N
495 Quadrangle Quadrangle No N/A N Y
496 Quarry Quar No N/A N Y
686 Quarters Quarters No N/A N Y
497 Quebrada Qbda Yes Creek Y N
499 Race Race No N/A N Y
501 Rail Rail No N/A N Y
502 Rail Link Rail Link No N/A Y Y
504 Railnet Railnet No N/A N Y
505 Railroad RR No N/A N Y
506 Railway Rlwy No N/A N Y
507 Ramal Ramal Yes Short Street Y N
508 Ramp Ramp No N/A N Y
510 Ranch Road Ranch Rd No N/A Y N
511 Ranch to Market Road RM No N/A Y N
512 Rancho Rch Yes Ranch or Farm Y N
513 Ravine Ravine No N/A N Y
514 Recreation Area Rec Area No N/A N Y
515 Reformatory Reformatory No N/A N Y
516 Refuge Refuge No N/A N Y
518 Regional Park Regional Pk No N/A N Y
519 Reservation Reservation No N/A N Y
520 Reservation Highway Resvn Hwy No N/A Y N
521 Reserve Resv No N/A N Y
522 Reservoir Reservoir No N/A Y Y
524 Residence Hall Res Hall No N/A N Y
525 Residencial Residencial Yes Public Housing Project Y N
526 Resort Resrt No N/A N Y
688 Rest Home Rest Home No N/A N Y
527 Retirement Home Retirement Hme No N/A N Y
528 Retirement Village Retirement Vlg No N/A N Y
529 Ridge Rdg No N/A N Y
543 Rio Río Yes River Y N
530 River Riv No N/A N Y
531 Road Rd No N/A Y Y
533 Roadway Roadway No N/A N Y
535 Rock Rock No N/A Y Y
536 Rooming House Rooming Hse No N/A N Y
537 Route Rte No N/A Y Y
538 Row Row No N/A Y Y
539 Rue Rue No N/A Y Y
540 Run Run No N/A N Y
541 Runway Runway No N/A Y Y
542 Ruta Ruta Yes Route Y N
498 RV Park RV Park No N/A N Y
545 Sanitarium Sanitarium No N/A N Y
546 School Schl No N/A Y Y
549 Sea Sea No N/A Y Y
550 Seashore Seashore No N/A N Y
552 Sector Sec Yes Sector Y N
553 Seminary Smry No N/A Y Y
554 Sendero Sendero Yes Foot Path Y N
555 Service Road Svc Rd No N/A Y Y
556 Shelter Shelter No N/A N Y
558 Shop Shop No N/A N Y
699 Shopping Center Shopping Ctr No N/A N Y
560 Shopping Mall Shopping Mall No N/A N Y
700 Shopping Plaza Shopping Plz No N/A N Y
703 Site Site No N/A N Y
564 Skyway Skwy No N/A Y Y
565 Slough Slough No N/A N Y
566 Sonda Sonda Yes Sound Y N
567 Sorority Sorority No N/A Y Y
568 Sound Snd No N/A Y N
569 Spa Spa No N/A Y Y
570 Speedway Speedway No N/A Y Y
571 Spring Spg No N/A N Y
572 Spur Spur No N/A Y Y
573 Square Sq No N/A Y Y
575 State Beach State Beach No N/A N Y
577 State Forest State Forest No N/A N Y
578 State Forest Service Road St FS Rd No N/A Y N
579 State Highway State Hwy No N/A Y N
580 State Hospital State Hospital No N/A Y Y
581 State Loop State Loop No N/A Y N
582 State Park State Park No N/A N Y
584 State Prison State Prison No N/A N Y
585 State Road State Rd No N/A Y N
586 State Route State Rte No N/A Y N
588 State Spur State Spur No N/A Y N
589 State Trunk Highway St Trunk Hwy No N/A Y N
591 Station Sta No N/A N Y
592 Strait Strait No N/A Y Y
593 Stravenue Stra No N/A N Y
594 Stream Strm No N/A N Y
595 Street St No N/A N Y
596 Strip Strip No N/A Y Y
599 Swamp Swamp No N/A N Y
600 Synagogue Synagogue No N/A Y Y
601 Tank Tank No N/A N Y
603 Temple Tmpl No N/A Y Y
604 Terminal Trmnl No N/A N Y
605 Terrace Ter No N/A Y Y
687 Thoroughfare Thoroughfare No N/A N Y
607 Toll Booth Toll Booth No N/A Y Y
701 Toll Road Toll Rd No N/A N Y
610 Tollway Tollway No N/A N Y
611 Tower Twr No N/A Y Y
612 Town Center Town Ctr No N/A Y Y
613 Town Hall Town Hall No N/A N Y
614 Town Highway Town Hwy No N/A Y N
615 Town Road Town Rd No N/A Y N
616 Towne Center Towne Ctr No N/A Y Y
617 Township Highway Twp Hwy No N/A Y N
618 Township Road Twp Rd No N/A Y N
619 Trace Trce No N/A N Y
620 Track Trak No N/A Y Y
621 Trafficway Trfy No N/A N Y
622 Trail Trl No N/A Y Y
623 Trailer Court Trailer Ct No N/A N Y
624 Trailer Park Trailer Pk No N/A N Y
628 Transmission Line Trans Ln No N/A N Y
702 Treatment Plant Trmt Plant No N/A Y Y
630 Tribal Road Tribal Rd No N/A Y N
632 Trolley Trolley No N/A Y Y
633 Truck Trail Truck Trl No N/A Y Y
636 Túnel Túnel Yes Tunnel Y N
634 Tunnel Tunl No N/A Y Y
635 Turnpike Tpke No N/A N Y
637 Underpass Upas No N/A Y Y
642 Universidad Universidad Yes University or College Y N
643 University Univ No N/A Y Y
638 US Forest Service Highway USFS Hwy No N/A Y N
639 US Forest Service Road USFS Rd No N/A Y N
640 US Highway US Hwy No N/A Y N
641 US Route US Rte No N/A Y N
644 Valley Vly No N/A N Y
645 Vereda Ver Yes Path Y N
655 Via Via Yes Way Y N
646 Viaduct Viaduct No N/A N Y
647 View Vw No N/A N Y
648 Villa Villa No N/A Y Y
649 Village Vlg No N/A Y Y
650 Village Center Village Ctr No N/A Y Y
697 Vineyard Vineyard No N/A N Y
652 Vineyards Vineyards No N/A N Y
654 Vista Vis Yes View Y Y
656 Walk Walk No N/A N Y
657 Walkway Walkway No N/A N Y
659 Wash Wash No N/A N Y
660 Waterway Waterway No N/A N Y
661 Way Way No N/A N Y
663 Wharf Wharf No N/A N Y
665 Wild and Scenic River Wld n Snc Riv No N/A N Y
664 Wild River Wild River No N/A N Y
666 Wilderness Wilderness No N/A N Y
667 Wilderness Park Wilderenss Pk No N/A N Y
668 Wildlife Management Area Wldlf Mgt Area No N/A N Y
669 Winery Winery No N/A Y Y
672 Yard Yard No N/A N Y
673 Yards Yards No N/A Y Y
670 YMCA YMCA No N/A Y Y
671 YWCA YWCA No N/A Y Y
675 Zanja Zanja Yes Ditch Y N
676 Zoo Zoo No N/A Y Y
*/

var featnames = []FeatnameInfo{
	{"103", "Academy", "Acdmy", false, "", true, true},
	{"104", "Acueducto", "Acueducto", true, "Aqueduct", true, false},
	{"105", "Aeropuerto", "Aero", true, "Airport", true, false},
	{"106", "Air Force Base", "AFB", false, "", false, true},
	{"107", "Airfield", "Airfield", false, "", false, true},
	{"108", "Airpark", "Airpark", false, "", false, true},
	{"109", "Airport", "Arprt", false, "", false, true},
	{"110", "Airstrip", "Airstrip", false, "", false, true},
	{"112", "Alley", "Aly", false, "", false, true},
	{"115", "Apartment Building", "Apt Bldg", false, "", false, true},
	{"116", "Apartment Complex", "Apt Complex", false, "", false, true},
	{"117", "Apartments", "Apts", false, "", false, true},
	{"118", "Aqueduct", "Aqueduct", false, "", false, true},
	{"119", "Arcade", "Arc", false, "", true, true},
	{"121", "Arroyo", "Arroyo", true, "Stream ", true, false},
	{"122", "Assisted Living Center", "Asstd Liv Ctr", false, "", false, true},
	{"694", "Assisted Living Facility", "Asstd Liv Fac", false, "", false, true},
	{"123", "Autopista", "Autopista", true, "Expressway or Freeway ", true, false},
	{"124", "Avenida", "Ave", true, "Avenue ", true, false},
	{"125", "Avenue", "Ave", false, "", true, true},
	{"126", "Bahia", "Bahía", true, "Bay ", true, false},
	{"127", "Bank", "Bk", false, "", true, true},
	{"704", "Base", "Base", false, "", false, true},
	{"128", "Basin", "Basin", false, "", false, true},
	{"129", "Bay", "Bay", false, "", true, true},
	{"130", "Bayou", "Byu", false, "", true, true},
	{"131", "Beach", "Bch", false, "", false, true},
	{"132", "Bed and Breakfast", "B and B", false, "", false, true},
	{"136", "Beltway", "Beltway", false, "", false, true},
	{"137", "Bend", "Bnd", false, "", false, true},
	{"138", "Bluff", "Blf", false, "", false, true},
	{"139", "Boarding House", "Brdng Hse", false, "", false, true},
	{"140", "Bog", "Bog", false, "", false, true},
	{"141", "Bosque", "Bosque", true, "Forest ", true, false},
	{"142", "Boulevard", "Blvd", false, "", true, true},
	{"143", "Boundary", "Boundary", false, "", false, true},
	{"146", "Branch", "Br", false, "", true, true},
	{"147", "Bridge", "Brg", false, "", false, true},
	{"148", "Brook", "Brk", false, "", false, true},
	{"149", "Building", "Bldg", false, "", true, true},
	{"150", "Bulevar", "Bulevar", true, "Boulevard ", true, false},
	{"151", "Bureau of Indian Affairs Highway", "BIA Hwy", false, "", true, false},
	{"152", "Bureau of Indian Affairs Road", "BIA Rd", false, "", true, false},
	{"153", "Bureau of Indian Affairs Route", "BIA Rte", false, "", true, false},
	{"154", "Bureau of Land Management Road", "BLM Rd", false, "", true, false},
	{"696", "Bypass", "Byp", false, "", true, true},
	{"156", "Calle", "Cll", true, "Street ", true, false},
	{"157", "Calleja", "Calleja", true, "Narrow Street ", true, false},
	{"158", "Callejón", "Callejón", true, "Alley ", true, false},
	{"159", "Caminito", "Cmt", true, "Little Road ", true, false},
	{"160", "Camino", "Cam", true, "Road or Way ", true, false},
	{"161", "Camp", "Cp", false, "", true, true},
	{"163", "Campground", "Cmpgrnd", false, "", false, true},
	{"164", "Campus", "Cmps", false, "", false, true},
	{"165", "Canal", "Cnl", false, "", true, true},
	{"172", "Cano", "Caño", true, "Drain or Sewer ", true, false},
	{"166", "Cantera", "Cantera", true, "Quarry or Gravel Pit", true, false},
	{"167", "Canyon", "Cyn", false, "", true, true},
	{"168", "Capilla", "Capilla", true, "Chapel ", true, false},
	{"169", "Carretera", "Carr", true, "Road ", true, false},
	{"170", "Causeway", "Cswy", false, "", false, true},
	{"171", "Cayo", "Cayo", true, "Key ", true, false},
	{"173", "Cementerio", "Cem", true, "Cemetery ", true, false},
	{"174", "Cemetery", "Cmtry", false, "", false, true},
	{"175", "Center", "Ctr", false, "", true, true},
	{"176", "Centro", "Centro", true, "Center ", true, false},
	{"177", "Cerrada", "Cer", true, "Closed ", true, false},
	{"178", "Chamber of Commerce", "Cham of Com", false, "", false, true},
	{"179", "Channel", "Chnnl", false, "", false, true},
	{"180", "Chapel", "Cpl", false, "", true, true},
	{"181", "Childrens Home", "Childrens Home", false, "", false, true},
	{"182", "Church", "Church", false, "", true, true},
	{"183", "Circle", "Cir", false, "", false, true},
	{"234", "Círculo", "Cír", true, "Circle ", true, false},
	{"184", "City Hall", "City Hall", false, "", false, true},
	{"185", "City Park", "City Park", false, "", false, true},
	{"186", "Cliff", "Clf", false, "", false, true},
	{"187", "Club", "Clb", false, "", true, true},
	{"188", "Colegio", "Colegio", true, "School ", true, false},
	{"189", "College", "Colg", false, "", true, true},
	{"190", "Common", "Cmn", false, "", false, true},
	{"191", "Commons", "Cmns", false, "", true, true},
	{"192", "Community Center", "Community Ctr", false, "", false, true},
	{"193", "Community College", "Community Colg", false, "", true, true},
	{"194", "Community Park", "Community Park", false, "", true, true},
	{"195", "Complex", "Complx", false, "", false, true},
	{"197", "Condominios", "Condios", true, "Condominiums ", true, false},
	{"198", "Condominium", "Condo", false, "", true, true},
	{"199", "Condominiums", "Condos", false, "", false, true},
	{"201", "Convent", "Cnvnt", false, "", true, true},
	{"202", "Convention Center", "Convention Ctr", false, "", true, true},
	{"203", "Corners", "Cors", false, "", false, true},
	{"204", "Correctional Facility", "Corr Faclty", false, "", false, true},
	{"205", "Correctional Institute", "Corr Inst", false, "", false, true},
	{"207", "Corte", "Corte", true, "Court ", true, false},
	{"679", "Cottage", "Cottage", false, "", false, true},
	{"208", "Coulee", "Coulee", false, "", false, true},
	{"209", "Country Club", "Country Club", false, "", true, true},
	{"210", "County Highway", "Co Hwy", false, "", true, false},
	{"211", "County Home", "Co Home", false, "", true, true},
	{"212", "County Lane", "Co Ln", false, "", true, false},
	{"213", "County Park", "Co Park", false, "", false, true},
	{"214", "County Road", "Co Rd", false, "", true, false},
	{"215", "County Route", "Co Rte", false, "", true, false},
	{"216", "County State Aid Highway", "Co St Aid Hwy", false, "", true, false},
	{"217", "County Trunk Highway", "Co Trunk Hwy", false, "", true, false},
	{"218", "County Trunk Road", "Co Trunk Rd", false, "", true, false},
	{"219", "Course", "Crs", false, "", false, true},
	{"220", "Court", "Ct", false, "", true, true},
	{"221", "Courthouse", "Courthouse", false, "", false, true},
	{"222", "Courts", "Cts", false, "", false, true},
	{"223", "Cove", "Cv", false, "", false, true},
	{"225", "Creek", "Crk", false, "", false, true},
	{"226", "Crescent", "Cres", false, "", false, true},
	{"227", "Crest", "Crst", false, "", false, true},
	{"228", "Crossing", "Xing", false, "", false, true},
	{"229", "Crossroads", "Xroad", false, "", true, true},
	{"233", "Cutoff", "Cutoff", false, "", false, true},
	{"235", "Dam", "Dm", false, "", false, true},
	{"236", "Delta Road", "Delta Rd", false, "", true, false},
	{"237", "Department", "Dept", false, "", true, true},
	{"238", "Depot", "Dep", false, "", false, true},
	{"239", "Detention Center", "Detention Ctr", false, "", false, true},
	{"240", "District of Columbia Highway", "DC Hwy", false, "", true, false},
	{"241", "Ditch", "Ditch", false, "", true, true},
	{"242", "Divide", "Dv", false, "", false, true},
	{"243", "Dock", "Dock", false, "", false, true},
	{"244", "Dormitory", "Dormitory", false, "", false, true},
	{"245", "Drain", "Drn", false, "", false, true},
	{"246", "Draw", "Draw", false, "", false, true},
	{"247", "Drive", "Dr", false, "", false, true},
	{"248", "Driveway", "Driveway", false, "", true, true},
	{"249", "Dump", "Dump", false, "", false, true},
	{"251", "Edificio", "Edif", true, "Building ", true, false},
	{"252", "Elementary School", "Elem School", false, "", false, true},
	{"253", "Ensenada", "Ensenada", true, "Cove ", true, false},
	{"254", "Entrada", "Ent", true, "Entrance ", true, false},
	{"256", "Escuela", "Escuela", true, "School ", true, false},
	{"680", "Esplanade", "Esplanade", true, "Esplanade", true, true},
	{"257", "Estates", "Ests", false, "", false, true},
	{"260", "Estuary", "Estuary", false, "", false, true},
	{"261", "Expreso", "Expreso", true, "Expressway ", true, false},
	{"262", "Expressway", "Expy", false, "", true, true},
	{"263", "Extension", "Ext", false, "", true, true},
	{"264", "Facility", "Faclty", false, "", false, true},
	{"265", "Fairgrounds", "Fairgrounds", false, "", false, true},
	{"266", "Falls", "Fls", false, "", true, true},
	{"267", "Farm", "Frm", false, "", false, true},
	{"268", "Farm Road", "Farm Rd", false, "", true, false},
	{"269", "Farm-to-Market", "Road FM", false, "", true, false},
	{"275", "Fence Line", "Fence Line", false, "", false, true},
	{"276", "Ferry Crossing", "Ferry Crossing", false, "", true, true},
	{"277", "Field", "Fld", false, "", false, true},
	{"278", "Fire Control Road", "Fire Cntrl Rd", false, "", true, false},
	{"279", "Fire Department", "Fire Dept", false, "", false, true},
	{"280", "Fire District Road", "Fire Dist Rd", false, "", true, false},
	{"281", "Fire Lane", "Fire Ln", false, "", true, false},
	{"282", "Fire Road", "Fire Rd", false, "", true, false},
	{"283", "Fire Route", "Fire Rte", false, "", true, false},
	{"284", "Fire Station", "Fire Sta", false, "", true, true},
	{"285", "Fire Trail", "Fire Trl", false, "", true, false},
	{"286", "Flowage", "Flowage", false, "", false, true},
	{"287", "Flume", "Flume", false, "", false, true},
	{"288", "Forest", "Frst", false, "", false, true},
	{"289", "Forest Highway", "Forest Hwy", false, "", true, true},
	{"290", "Forest Road", "Forest Rd", false, "", true, false},
	{"291", "Forest Route", "Forest Rte", false, "", true, false},
	{"292", "Forest Service Road", "FS Rd", false, "", true, false},
	{"293", "Fork", "Frk", false, "", false, true},
	{"294", "Fort", "Ft", false, "", true, false},
	{"295", "Four-Wheel Drive Trail", "4WD Trl", false, "", true, true},
	{"296", "Fraternity", "Frtrnty", false, "", false, true},
	{"297", "Freeway", "Fwy", false, "", false, true},
	{"298", "Garage", "Grge", false, "", false, true},
	{"299", "Gardens", "Gdns", false, "", false, true},
	{"303", "Glacier", "Glacier", false, "", false, true},
	{"304", "Glen", "Gln", false, "", false, true},
	{"305", "Golf Club", "Golf Club", false, "", true, true},
	{"306", "Golf Course", "Golf Course", false, "", true, true},
	{"307", "Grade", "Grade", false, "", false, true},
	{"309", "Green", "Grn", false, "", false, true},
	{"310", "Group Home", "Group Home", false, "", false, true},
	{"311", "Gulch", "Gulch", false, "", false, true},
	{"312", "Gulf", "Gulf", false, "", true, true},
	{"313", "Gully", "Gully", false, "", false, true},
	{"314", "Halfway House", "Halfway House", false, "", false, true},
	{"315", "Hall", "Hall", false, "", false, true},
	{"316", "Harbor", "Hbr", false, "", false, true},
	{"317", "Heights", "Hts", false, "", false, true},
	{"321", "High School", "High School", false, "", false, true},
	{"322", "Highway", "Hwy", false, "", true, true},
	{"323", "Hill", "Hl", false, "", false, true},
	{"324", "Hollow", "Holw", false, "", false, true},
	{"325", "Home", "Home", false, "", true, true},
	{"326", "Hospital", "Hosp", false, "", true, true},
	{"327", "Hostel", "Hostel", false, "", false, true},
	{"328", "Hotel", "Hotel", false, "", true, true},
	{"329", "House", "Hse", false, "", true, true},
	{"330", "Housing", "Hsng", false, "", true, true},
	{"332", "Iglesia", "Iglesia", true, "Church", true, false},
	{"333", "Indian Route", "Indian Rte", false, "", true, false},
	{"334", "Indian Service Route", "Indian Svc Rte", false, "", true, false},
	{"336", "Industrial Park", "Indl Park", false, "", false, true},
	{"337", "Inlet", "Inlt", false, "", false, true},
	{"338", "Inn", "Inn", false, "", true, true},
	{"339", "Institute", "Inst", false, "", true, true},
	{"340", "Institution", "Instn", false, "", false, true},
	{"341", "Instituto", "Instituto", true, "Institute", true, false},
	{"342", "Intermediate School", "Inter School", false, "", false, true},
	{"344", "Interstate Highway", "I-", false, "", true, false},
	{"345", "Isla", "Isla", true, "Island", true, false},
	{"346", "Island", "Is", false, "", false, true},
	{"347", "Islands", "Iss", false, "", true, true},
	{"348", "Isle", "Isle", false, "", true, true},
	{"349", "Jail", "Jail", false, "", false, true},
	{"351", "Jeep Trail", "Jeep Trl", false, "", true, true},
	{"352", "Junction", "Junction", false, "", false, true},
	{"353", "Junior High School", "Jr HS", false, "", false, true},
	{"356", "Kill", "Kill", false, "", true, true},
	{"357", "Lago", "Lago", true, "Lake ", true, false},
	{"358", "Lagoon", "Lagoon", false, "", false, true},
	{"360", "Laguna", "Laguna", true, "Lagoon ", true, false},
	{"361", "Lake", "Lk", false, "", true, true},
	{"362", "Lakes", "Lks", false, "", false, true},
	{"363", "Landfill", "Lndfll", false, "", false, true},
	{"364", "Landing", "Lndg", false, "", false, true},
	{"365", "Landing Area", "Landing Area", false, "", true, true},
	{"366", "Landing Field", "Landing Fld", false, "", true, true},
	{"367", "Landing Strip", "Landing Strp", false, "", true, true},
	{"368", "Lane", "Ln", false, "", true, true},
	{"369", "Lateral", "Lateral", false, "", true, true},
	{"370", "Levee", "Levee", false, "", true, true},
	{"371", "Library", "Lbry", false, "", true, true},
	{"372", "Lift", "Lift", false, "", true, true},
	{"373", "Lighthouse", "Lighthouse", false, "", false, true},
	{"374", "Line", "Line", false, "", true, true},
	{"376", "Lodge", "Ldg", false, "", false, true},
	{"377", "Logging Road", "Logging Rd", false, "", true, true},
	{"378", "Loop", "Loop", false, "", true, true},
	{"379", "Mall", "Mall", false, "", true, true},
	{"380", "Manor", "Mnr", false, "", false, true},
	{"381", "Mar", "Mar", true, "Sea ", true, false},
	{"382", "Marginal", "Marginal", true, "Service Road ", true, false},
	{"383", "Marina", "Mrna", false, "", false, true},
	{"384", "Marsh", "Marsh", false, "", false, true},
	{"385", "Meadows", "Mdws", false, "", false, true},
	{"386", "Medical Building", "Medical Bldg", false, "", false, true},
	{"387", "Medical Center", "Medical Ctr", false, "", true, true},
	{"388", "Memorial", "Meml", false, "", false, true},
	{"389", "Memorial Gardens", "Memorial Gnds", false, "", false, true},
	{"390", "Memorial Park", "Memorial Pk", false, "", false, true},
	{"391", "Mesa", "Mesa", false, "", true, true},
	{"392", "Middle School", "Mid Schl", false, "", false, true},
	{"393", "Military Reservation", "Mil Res", false, "", false, true},
	{"394", "Millpond", "Millpond", false, "", false, true},
	{"395", "Mine", "Mine", false, "", false, true},
	{"396", "Mission", "Mssn", false, "", true, true},
	{"397", "Mobile Home Community", "Mobile Hm Cmty", false, "", true, true},
	{"398", "Mobile Home Estates", "Mobile Hm Est", false, "", true, true},
	{"399", "Mobile Home Park", "Mobile Hm Pk", false, "", true, true},
	{"400", "Monastery", "Monstry", false, "", true, true},
	{"401", "Monument", "Mnmt", false, "", false, true},
	{"403", "Mosque", "Mosque", false, "", true, true},
	{"404", "Motel", "Mtl", false, "", true, true},
	{"405", "Motor Lodge", "Motor Lodge", false, "", false, true},
	{"406", "Motorway", "Mtwy", false, "", false, true},
	{"407", "Mount", "Mt", false, "", true, true},
	{"408", "Mountain", "Mtn", false, "", false, true},
	{"411", "Museum", "Mus", false, "", true, true},
	{"412", "National Battlefield", "Natl Bfld", false, "", false, true},
	{"413", "National Battlefield Park", "Natl Bfld Pk", false, "", false, true},
	{"414", "National Battlefield Site", "Natl Bfld Site", false, "", false, true},
	{"415", "National Conservation Area", "Natl Cnsv Area", false, "", false, true},
	{"416", "National Forest", "Natl Forest", false, "", false, true},
	{"417", "National Forest Development Road", "Nat For Dev Rd", false, "", true, false},
	{"419", "National Grasslands", "Natl Grsslnds", false, "", false, true},
	{"420", "National Historic Site", "Natl Hist Site", false, "", false, true},
	{"421", "National Historical Park", "Natl Hist Pk", false, "", false, true},
	{"422", "National Lakeshore", "Natl Lkshr", false, "", false, true},
	{"423", "National Memorial", "Natl Meml", false, "", false, true},
	{"424", "National Military Park", "Natl Mil Pk", false, "", false, true},
	{"425", "National Monument", "Natl Mnmt", false, "", false, true},
	{"426", "National Park", "Natl Pk", false, "", false, true},
	{"427", "National Preserve", "Natl Prsv", false, "", false, true},
	{"428", "National Recreation Area", "Natl Rec Area", false, "", false, true},
	{"429", "National Recreational River", "Natl Rec Riv", false, "", false, true},
	{"430", "National Reserve", "Natl Resv", false, "", false, true},
	{"431", "National River", "Natl Riv", false, "", false, true},
	{"432", "National Scenic Area", "Natl Sc Area", false, "", false, true},
	{"433", "National Scenic River", "Natl Sc Riv", false, "", false, true},
	{"435", "National Scenic Riverways", "Natl Sc Rvrwys", false, "", false, true},
	{"436", "National Scenic Trail", "Natl Sc Trl", false, "", false, true},
	{"437", "National Seashore", "Natl Shr", false, "", false, true},
	{"438", "National Wildlife Refuge", "Natl Wld Rfg", false, "", false, true},
	{"439", "Navajo Service Route", "Navajo Svc Rte", false, "", true, false},
	{"440", "Naval Air Station", "Naval Air Sta", false, "", false, true},
	{"442", "Nursing Home", "Nurse Home", false, "", false, true},
	{"444", "Ocean", "Ocean", false, "", false, true},
	{"445", "Oceano", "Océano", true, "Ocean ", true, false},
	{"446", "Office", "Ofc", false, "", true, true},
	{"447", "Office Building", "Office Bldg", false, "", false, true},
	{"449", "Office Park", "Office Park", false, "", false, true},
	{"698", "Orchard", "Orchard", false, "", false, true},
	{"451", "Orchards", "Orchrds", false, "", false, true},
	{"452", "Orphanage", "Orphanage", false, "", false, true},
	{"453", "Outlet", "Outlet", false, "", false, true},
	{"454", "Oval", "Oval", false, "", false, true},
	{"455", "Overpass", "Opas", false, "", false, true},
	{"456", "Parish Road", "Parish Rd", false, "", true, false},
	{"457", "Park", "Park", false, "", false, true},
	{"458", "Park and Ride", "Park and Ride", false, "", false, true},
	{"460", "Parkway", "Pkwy", false, "", false, true},
	{"706", "Parq", "Parq", true, "Park ", true, false},
	{"461", "Parque", "Parque", true, "Park ", true, false},
	{"462", "Pasaje", "Pasaje", true, "Passage ", true, false},
	{"463", "Paseo", "Pso", true, "Path ", true, false},
	{"464", "Pass", "Pass", false, "", true, true},
	{"465", "Passage", "Psge", false, "", true, true},
	{"466", "Path", "Path", false, "", false, true},
	{"682", "Pavilion", "Pavilion", false, "", false, true},
	{"467", "Peak", "Peak", false, "", false, true},
	{"705", "Penitentiary", "Penitentiary", false, "", false, true},
	{"468", "Pier", "Pier", false, "", true, true},
	{"469", "Pike", "Pike", false, "", false, true},
	{"470", "Pipeline", "Pipeline", false, "", false, true},
	{"472", "Place", "Pl", false, "", false, true},
	{"473", "Placita", "Pla", true, "Little Plaza ", true, false},
	{"474", "Plant", "Plnt", false, "", false, true},
	{"683", "Plantation", "Plantation", false, "", false, true},
	{"475", "Playa", "Playa", true, "Beach ", true, false},
	{"476", "Playground", "Playground", false, "", false, true},
	{"477", "Plaza", "Plz", false, "", true, true},
	{"478", "Point", "Pt", false, "", true, true},
	{"479", "Pointe", "Pointe", false, "", false, true},
	{"480", "Police Department", "Police Dept", false, "", true, true},
	{"481", "Police Station", "Police Station", false, "", true, true},
	{"482", "Pond", "Pond", false, "", true, true},
	{"483", "Ponds", "Ponds", false, "", false, true},
	{"485", "Port", "Prt", false, "", true, true},
	{"486", "Post Office", "Post Office", false, "", false, true},
	{"487", "Power Line", "Power Line", false, "", false, true},
	{"691", "Power Plant", "Power Plant", false, "", false, true},
	{"488", "Prairie", "Pr", false, "", false, true},
	{"489", "Preserve", "Preserve", false, "", false, true},
	{"491", "Prison", "Prison", false, "", false, true},
	{"690", "Prison Farm", "Prison Farm", false, "", false, true},
	{"685", "Promenade", "Promenade", false, "", false, true},
	{"492", "Prong", "Prong", false, "", false, true},
	{"494", "Puente", "Puente", true, "Bridge ", true, false},
	{"495", "Quadrangle", "Quadrangle", false, "", false, true},
	{"496", "Quarry", "Quar", false, "", false, true},
	{"686", "Quarters", "Quarters", false, "", false, true},
	{"497", "Quebrada", "Qbda", true, "Creek ", true, false},
	{"499", "Race", "Race", false, "", false, true},
	{"501", "Rail", "Rail", false, "", false, true},
	{"502", "Rail Link", "Rail Link", false, "", true, true},
	{"504", "Railnet", "Railnet", false, "", false, true},
	{"505", "Railroad", "RR", false, "", false, true},
	{"506", "Railway", "Rlwy", false, "", false, true},
	{"507", "Ramal", "Ramal", true, "Short Street ", true, false},
	{"508", "Ramp", "Ramp", false, "", false, true},
	{"510", "Ranch Road", "Ranch Rd", false, "", true, false},
	{"511", "Ranch to Market Road", "RM", false, "", true, false},
	{"512", "Rancho", "Rch", true, "Ranch or Farm ", true, false},
	{"513", "Ravine", "Ravine", false, "", false, true},
	{"514", "Recreation Area", "Rec Area", false, "", false, true},
	{"515", "Reformatory", "Reformatory", false, "", false, true},
	{"516", "Refuge", "Refuge", false, "", false, true},
	{"518", "Regional Park", "Regional Pk", false, "", false, true},
	{"519", "Reservation", "Reservation", false, "", false, true},
	{"520", "Reservation Highway", "Resvn Hwy", false, "", true, false},
	{"521", "Reserve", "Resv", false, "", false, true},
	{"522", "Reservoir", "Reservoir", false, "", true, true},
	{"524", "Residence Hall", "Res Hall", false, "", false, true},
	{"525", "Residencial", "Residencial", true, "Public Housing Project ", true, false},
	{"526", "Resort", "Resrt", false, "", false, true},
	{"688", "Rest Home", "Rest Home", false, "", false, true},
	{"527", "Retirement Home", "Retirement Hme", false, "", false, true},
	{"528", "Retirement Village", "Retirement Vlg", false, "", false, true},
	{"529", "Ridge", "Rdg", false, "", false, true},
	{"543", "Rio", "Río", true, "River ", true, false},
	{"530", "River", "Riv", false, "", false, true},
	{"531", "Road", "Rd", false, "", true, true},
	{"533", "Roadway", "Roadway", false, "", false, true},
	{"535", "Rock", "Rock", false, "", true, true},
	{"536", "Rooming House", "Rooming Hse", false, "", false, true},
	{"537", "Route", "Rte", false, "", true, true},
	{"538", "Row", "Row", false, "", true, true},
	{"539", "Rue", "Rue", false, "", true, true},
	{"540", "Run", "Run", false, "", false, true},
	{"541", "Runway", "Runway", false, "", true, true},
	{"542", "Ruta", "Ruta", true, "Route ", true, false},
	{"498", "RV Park", "RV Park", false, "", false, true},
	{"545", "Sanitarium", "Sanitarium", false, "", false, true},
	{"546", "School", "Schl", false, "", true, true},
	{"549", "Sea", "Sea", false, "", true, true},
	{"550", "Seashore", "Seashore", false, "", false, true},
	{"552", "Sector", "Sec", true, "Sector ", true, false},
	{"553", "Seminary", "Smry", false, "", true, true},
	{"554", "Sendero", "Sendero", true, "Foot Path ", true, false},
	{"555", "Service Road", "Svc Rd", false, "", true, true},
	{"556", "Shelter", "Shelter", false, "", false, true},
	{"558", "Shop", "Shop", false, "", false, true},
	{"699", "Shopping Center", "Shopping Ctr", false, "", false, true},
	{"560", "Shopping Mall", "Shopping Mall", false, "", false, true},
	{"700", "Shopping Plaza", "Shopping Plz", false, "", false, true},
	{"703", "Site", "Site", false, "", false, true},
	{"564", "Skyway", "Skwy", false, "", true, true},
	{"565", "Slough", "Slough", false, "", false, true},
	{"566", "Sonda", "Sonda", true, "Sound ", true, false},
	{"567", "Sorority", "Sorority", false, "", true, true},
	{"568", "Sound", "Snd", false, "", true, false},
	{"569", "Spa", "Spa", false, "", true, true},
	{"570", "Speedway", "Speedway", false, "", true, true},
	{"571", "Spring", "Spg", false, "", false, true},
	{"572", "Spur", "Spur", false, "", true, true},
	{"573", "Square", "Sq", false, "", true, true},
	{"575", "State Beach", "State Beach", false, "", false, true},
	{"577", "State Forest", "State Forest", false, "", false, true},
	{"578", "State Forest Service Road", "St FS Rd", false, "", true, false},
	{"579", "State Highway", "State Hwy", false, "", true, false},
	{"580", "State Hospital", "State Hospital", false, "", true, true},
	{"581", "State Loop", "State Loop", false, "", true, false},
	{"582", "State Park", "State Park", false, "", false, true},
	{"584", "State Prison", "State Prison", false, "", false, true},
	{"585", "State Road", "State Rd", false, "", true, false},
	{"586", "State Route", "State Rte", false, "", true, false},
	{"588", "State Spur", "State Spur", false, "", true, false},
	{"589", "State Trunk Highway", "St Trunk Hwy", false, "", true, false},
	{"591", "Station", "Sta", false, "", false, true},
	{"592", "Strait", "Strait", false, "", true, true},
	{"593", "Stravenue", "Stra", false, "", false, true},
	{"594", "Stream", "Strm", false, "", false, true},
	{"595", "Street", "St", false, "", false, true},
	{"596", "Strip", "Strip", false, "", true, true},
	{"599", "Swamp", "Swamp", false, "", false, true},
	{"600", "Synagogue", "Synagogue", false, "", true, true},
	{"601", "Tank", "Tank", false, "", false, true},
	{"603", "Temple", "Tmpl", false, "", true, true},
	{"604", "Terminal", "Trmnl", false, "", false, true},
	{"605", "Terrace", "Ter", false, "", true, true},
	{"687", "Thoroughfare", "Thoroughfare", false, "", false, true},
	{"607", "Toll Booth", "Toll Booth", false, "", true, true},
	{"701", "Toll Road", "Toll Rd", false, "", false, true},
	{"610", "Tollway", "Tollway", false, "", false, true},
	{"611", "Tower", "Twr", false, "", true, true},
	{"612", "Town Center", "Town Ctr", false, "", true, true},
	{"613", "Town Hall", "Town Hall", false, "", false, true},
	{"614", "Town Highway", "Town Hwy", false, "", true, false},
	{"615", "Town Road", "Town Rd", false, "", true, false},
	{"616", "Towne Center", "Towne Ctr", false, "", true, true},
	{"617", "Township Highway", "Twp Hwy", false, "", true, false},
	{"618", "Township Road", "Twp Rd", false, "", true, false},
	{"619", "Trace", "Trce", false, "", false, true},
	{"620", "Track", "Trak", false, "", true, true},
	{"621", "Trafficway", "Trfy", false, "", false, true},
	{"622", "Trail", "Trl", false, "", true, true},
	{"623", "Trailer Court", "Trailer Ct", false, "", false, true},
	{"624", "Trailer Park", "Trailer Pk", false, "", false, true},
	{"628", "Transmission Line", "Trans Ln", false, "", false, true},
	{"702", "Treatment Plant", "Trmt Plant", false, "", true, true},
	{"630", "Tribal Road", "Tribal Rd", false, "", true, false},
	{"632", "Trolley", "Trolley", false, "", true, true},
	{"633", "Truck Trail", "Truck Trl", false, "", true, true},
	{"636", "Túnel", "Túnel", true, "Tunnel ", true, false},
	{"634", "Tunnel", "Tunl", false, "", true, true},
	{"635", "Turnpike", "Tpke", false, "", false, true},
	{"637", "Underpass", "Upas", false, "", true, true},
	{"642", "Universidad", "Universidad", true, "University or College ", true, false},
	{"643", "University", "Univ", false, "", true, true},
	{"638", "US Forest Service Highway", "USFS Hwy", false, "", true, false},
	{"639", "US Forest Service Road", "USFS Rd", false, "", true, false},
	{"640", "US Highway", "US Hwy", false, "", true, false},
	{"641", "US Route", "US Rte", false, "", true, false},
	{"644", "Valley", "Vly", false, "", false, true},
	{"645", "Vereda", "Ver", true, "Path ", true, false},
	{"655", "Via", "Via", true, "Way ", true, false},
	{"646", "Viaduct", "Viaduct", false, "", false, true},
	{"647", "View", "Vw", false, "", false, true},
	{"648", "Villa", "Villa", false, "", true, true},
	{"649", "Village", "Vlg", false, "", true, true},
	{"650", "Village Center", "Village Ctr", false, "", true, true},
	{"697", "Vineyard", "Vineyard", false, "", false, true},
	{"652", "Vineyards", "Vineyards", false, "", false, true},
	{"654", "Vista", "Vis", true, "View", true, true},
	{"656", "Walk", "Walk", false, "", false, true},
	{"657", "Walkway", "Walkway", false, "", false, true},
	{"659", "Wash", "Wash", false, "", false, true},
	{"660", "Waterway", "Waterway", false, "", false, true},
	{"661", "Way", "Way", false, "", false, true},
	{"663", "Wharf", "Wharf", false, "", false, true},
	{"665", "Wild and Scenic River", "Wld n Snc Riv", false, "", false, true},
	{"664", "Wild River", "Wild River", false, "", false, true},
	{"666", "Wilderness", "Wilderness", false, "", false, true},
	{"667", "Wilderness Park", "Wilderenss Pk", false, "", false, true},
	{"668", "Wildlife Management Area", "Wldlf Mgt Area", false, "", false, true},
	{"669", "Winery", "Winery", false, "", true, true},
	{"672", "Yard", "Yard", false, "", false, true},
	{"673", "Yards", "Yards", false, "", true, true},
	{"670", "YMCA", "YMCA", false, "", true, true},
	{"671", "YWCA", "YWCA", false, "", true, true},
	{"675", "Zanja", "Zanja", true, "Ditch ", true, false},
	{"676", "Zoo", "Zoo", false, "", true, true},
}

var featnameMap = maps.Collect(func(yield func(string, FeatnameInfo) bool) {
	for _, v := range featnames {
		if !yield(v.Code, v) {
			return
		}
	}
})

func ExpandFeatureName(attr map[string]any) string {
	isSpanish := false

	base := ""
	// attr['NAME'] will contain the text between all prefix and suffix values
	rawname, ok := attr["NAME"]
	if ok {
		base = fieldutil.AsString(rawname)
	}

	prefixqualifier := ""
	// attr['PREQUAL'] will contain a numeric code for qualifiers
	rawpq, ok := attr["PREQUAL"]
	if ok {
		pq := fieldutil.AsString(rawpq)
		prefixQualifierInfo, err := qualifiers.Info(pq)
		if err == nil && prefixQualifierInfo.Prefix {
			prefixqualifier = prefixQualifierInfo.Full
		}
	}

	prefixtype := ""
	// attr['PRETYP'] will contain a numeric code for feature names
	pt := ""
	rawpt, ok := attr["PRETYP"]
	if ok {
		pt = fieldutil.AsString(rawpt)
	}
	prefixInfo, ok := featnameMap[pt]
	if ok && prefixInfo.Prefix {
		if prefixInfo.Spanish {
			isSpanish = true
		}
		prefixtype = prefixInfo.Full
	}

	suffixqualifier := ""
	// attr['SUFQUAL'] will contain a numeric code for qualifiers
	rawsq, ok := attr["SUFQUAL"]
	if ok {
		sq := fieldutil.AsString(rawsq)
		suffixQualifierInfo, err := qualifiers.Info(sq)
		if err == nil && suffixQualifierInfo.Suffix {
			suffixqualifier = suffixQualifierInfo.Full
		}
	}

	suffixtype := ""
	// attr['SUFTYP'] will contain a numeric code for feature names
	st := ""
	rawst, ok := attr["SUFTYP"]
	if ok {
		st = fieldutil.AsString(rawst)
	}
	suffixInfo, ok := featnameMap[st]
	if ok && suffixInfo.Suffix {
		if suffixInfo.Spanish {
			isSpanish = true
		}
		suffixtype = suffixInfo.Full
	}

	// handle directionals

	prefixdirectional := ""
	// attr['PREDIRABRV'] will contain a string abreviation for any predirectional
	rawpdir, ok := attr["PREDIRABRV"]
	if ok {
		pdir := fieldutil.AsString(rawpdir)
		if len(pdir) > 0 {
			pd := directionals.Expand(pdir, isSpanish)
			prefixdirectional = pd
		}
	}

	suffixdirectional := ""
	// attr['SUFDIRABRV'] will contain a string abreviation for any postdirectional
	rawsdir, ok := attr["SUFDIRABRV"]
	if ok {
		sdir := fieldutil.AsString(rawsdir)
		if len(sdir) > 0 {
			// FIXME: Exactly what we do with directionals is TBD.
			// See https://github.com/poetic-systems/zipcity/issues/4.
			// Expanding, abbreviating, and aliasing (duplicating without the directional)
			// all have significant concerns.
			sd := directionals.Expand(sdir, isSpanish)
			suffixdirectional = sd
		}
	}

	// According to USPS Pub 28 Puerto Rican addresses begin with the street type.
	// Additionally, directional prefixes are noted as rare (and "Ó " is not a
	// valid directional prefix.)
	// https://www2.census.gov/geo/tiger/rd_2ktiger/tgrrd2k.pdf lists "Ó" as one
	// of the characters that it previously used square brackets to indicate.
	// On that basis, the large number of streets in puerto rico starting with
	// "Ó " are believed to result from the migration of pre-2000 ASCII-to-UTF-8
	// diacritical encodings in Puerto Rican/Spanish street records, which persist
	// as literal strings in annual TIGER/Line roll-forwards.
	base, _ = strings.CutPrefix(base, "Ó ")

	// TODO: determine if we need to uppercase and remove diacritics preemptively

	// The full concatenation order in TIGER files is:
	//   Prefix Qualifier (e.g., Old, New)
	//   Prefix Directional (e.g., North, East)
	//   Prefix Type (e.g., State Route, County Road)
	//   Base Name
	//   Suffix Type
	//   Suffix Directional
	//   Suffix Qualifier

	return strings.Join([]string{
		prefixqualifier,
		prefixdirectional,
		prefixtype,
		base,
		suffixtype,
		suffixdirectional,
		suffixqualifier,
	}, " ")
}
