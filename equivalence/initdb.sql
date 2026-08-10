-- SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
--
-- SPDX-License-Identifier: CC0-1.0

-- Seeds the real C# app's Postgres config tables to match
-- src/config.example.yaml's subtypes/recipients exactly (same
-- category->classname/typeValue/severity mappings, same single IT492
-- recipient, "prohibition" disabled), so any output difference between the
-- two implementations is due to the code, not divergent configuration.

SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';
SET default_table_access_method = heap;

CREATE TABLE public."TAB_DESTINATARI" (
    "IdDestinatario" integer NOT NULL,
    "Descrizione" character varying(200) NOT NULL,
    "Invio" boolean NOT NULL,
    "Ricezione" boolean NOT NULL,
    "SupplierInterno" character varying(20) NOT NULL,
    "Supplier" character varying(20) NOT NULL,
    "Traduttore" character varying(200) NOT NULL,
    "Confidentiality" character varying(200) NOT NULL,
    "Abilitato" boolean NOT NULL,
    "Url" character varying(1000)
);

CREATE TABLE public."TAB_PARAMETRI" (
    "IdParametro" integer NOT NULL,
    "Modulo" character varying(100) NOT NULL,
    "Parametro" character varying(300) NOT NULL,
    "Valore" character varying(4000),
    "Note" character varying(1000)
);

CREATE TABLE public."TAB_SOTTOTIPI" (
    "IdSottotipo" integer NOT NULL,
    "Descrizione" character varying(200) NOT NULL,
    "TypeCode" character varying(200) NOT NULL,
    "SubtypeCode" character varying(200) NOT NULL,
    "Classname" character varying(200) NOT NULL,
    "TypeValue" character varying(200) NOT NULL,
    "ExtraAttribute" character varying(200),
    "ExtraValue" character varying(200),
    "Ingresso" boolean NOT NULL,
    "Uscita" boolean NOT NULL,
    "Abilitato" boolean NOT NULL,
    "Severity" character varying(50)
);

COPY public."TAB_DESTINATARI" ("IdDestinatario", "Descrizione", "Invio", "Ricezione", "SupplierInterno", "Supplier", "Traduttore", "Confidentiality", "Abilitato", "Url") FROM stdin;
1	A22 - Brenner motorway	t	f	ITPBZ	IT492	DatexOrig	1,2,3,4	t	\N
\.

COPY public."TAB_PARAMETRI" ("IdParametro", "Modulo", "Parametro", "Valore", "Note") FROM stdin;
1	NodoDatex	Installazione	BZ	\N
2	NodoDatex	SupplierInterno	ITPBZ	\N
\.

COPY public."TAB_SOTTOTIPI" ("IdSottotipo", "Descrizione", "TypeCode", "SubtypeCode", "Classname", "TypeValue", "ExtraAttribute", "ExtraValue", "Ingresso", "Uscita", "Abilitato", "Severity") FROM stdin;
7	Event	GENERIC	traffic-event:event	PublicEvent	majorEvent	\N	\N	f	t	t	medium
8	Prohibition	GENERIC	traffic-event:prohibition	RoadOrCarriagewayOrLaneManagement	other	\N	\N	f	f	f	high
5	Road work	GENERIC	traffic-event:road-work	MaintenanceWorks	roadworks	\N	\N	f	t	t	medium
2	Accident	GENERIC	traffic-event:accident	Accident	accident	\N	\N	f	t	t	high
3	Animal on road	GENERIC	traffic-event:animal-on-road	AnimalPresenceObstruction	animalsOnTheRoad	\N	\N	f	t	t	high
9	Closure	GENERIC	traffic-event:closure	RoadOrCarriagewayOrLaneManagement	roadClosed	\N	\N	f	t	t	high
10	Hindrance	GENERIC	traffic-event:hindrance	RoadOrCarriagewayOrLaneManagement	singleAlternateLineTraffic	\N	\N	f	t	t	medium
4	Caution	GENERIC	traffic-event:caution	RoadOrCarriagewayOrLaneManagement	other	\N	\N	f	t	t	medium
1	Congestion	GENERIC	traffic-event:congestion	AbnormalTraffic	slowTraffic	\N	\N	f	t	t	high
6	Weather related	GENERIC	traffic-event:weather-related	PoorEnvironmentCondition	fog	\N	\N	f	t	t	high
11	Speed camera	GENERIC	traffic-event:speed-camera	SpeedManagement	policeSpeedChecksInOperation	\N	\N	f	t	t	low
12	Road condition	GENERIC	traffic-event:road-condition	WinterDrivingManagement	winterEquipmentOnBoardRequired	complianceOption	mandatory	f	t	t	medium
\.

ALTER TABLE ONLY public."TAB_DESTINATARI"
    ADD CONSTRAINT "TAB_DESTINATARI_pkey" PRIMARY KEY ("IdDestinatario");

ALTER TABLE ONLY public."TAB_PARAMETRI"
    ADD CONSTRAINT "TAB_PARAMETRI_pkey" PRIMARY KEY ("IdParametro");

ALTER TABLE ONLY public."TAB_SOTTOTIPI"
    ADD CONSTRAINT "TAB_SOTTOTIPI_pkey" PRIMARY KEY ("IdSottotipo");
