--
-- PostgreSQL database dump
--

-- Dumped from database version 15.8
-- Dumped by pg_dump version 15.8

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: addresses; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.addresses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    number bigint NOT NULL,
    route text NOT NULL,
    optional_route text,
    city text NOT NULL,
    zip_code text NOT NULL,
    country text NOT NULL,
    latitude numeric NOT NULL,
    longitude numeric NOT NULL
);


ALTER TABLE public.addresses OWNER TO balade;

--
-- Name: addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.addresses_id_seq OWNER TO balade;

--
-- Name: addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.addresses_id_seq OWNED BY public.addresses.id;


--
-- Name: guides; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.guides (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text NOT NULL,
    phone text,
    bio text,
    experience text,
    specialties text,
    languages text,
    certification_level text,
    avatar text,
    is_active boolean DEFAULT true NOT NULL,
    emergency_contact_name text,
    emergency_contact_phone text,
    stripe_account_id text,
    stripe_public_key text,
    stripe_secret_key text,
    stripe_webhook_secret text,
    payment_enabled boolean DEFAULT false NOT NULL,
    user_id bigint
);


ALTER TABLE public.guides OWNER TO balade;

--
-- Name: guides_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.guides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.guides_id_seq OWNER TO balade;

--
-- Name: guides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.guides_id_seq OWNED BY public.guides.id;


--
-- Name: payments; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.payments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    stripe_payment_intent_id text NOT NULL,
    stripe_charge_id text,
    amount bigint NOT NULL,
    currency text DEFAULT 'eur'::text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    payment_method text DEFAULT 'card'::text,
    failure_reason text,
    registration_id bigint,
    group_id bigint,
    payer_email text NOT NULL,
    payer_name text NOT NULL,
    guide_id bigint NOT NULL,
    paid_at timestamp with time zone,
    refunded_at timestamp with time zone,
    refund_amount bigint
);


ALTER TABLE public.payments OWNER TO balade;

--
-- Name: payments_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.payments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.payments_id_seq OWNER TO balade;

--
-- Name: payments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.payments_id_seq OWNED BY public.payments.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    readable_name text,
    description text,
    category text
);


ALTER TABLE public.permissions OWNER TO balade;

--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.permissions_id_seq OWNER TO balade;

--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: ramble_guides; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.ramble_guides (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ramble_id bigint NOT NULL,
    guide_id bigint NOT NULL
);


ALTER TABLE public.ramble_guides OWNER TO balade;

--
-- Name: ramble_guides_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.ramble_guides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ramble_guides_id_seq OWNER TO balade;

--
-- Name: ramble_guides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.ramble_guides_id_seq OWNED BY public.ramble_guides.id;


--
-- Name: ramble_prices; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.ramble_prices (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    label text NOT NULL,
    amount numeric NOT NULL,
    ramble_id bigint NOT NULL
);


ALTER TABLE public.ramble_prices OWNER TO balade;

--
-- Name: ramble_prices_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.ramble_prices_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ramble_prices_id_seq OWNER TO balade;

--
-- Name: ramble_prices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.ramble_prices_id_seq OWNED BY public.ramble_prices.id;


--
-- Name: ramble_registration_groups; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.ramble_registration_groups (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    primary_email text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    registration_date timestamp with time zone NOT NULL,
    confirmation_date timestamp with time zone,
    confirmation_deadline timestamp with time zone,
    cancellation_date timestamp with time zone,
    cancellation_reason text,
    ramble_id bigint NOT NULL
);


ALTER TABLE public.ramble_registration_groups OWNER TO balade;

--
-- Name: ramble_registration_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.ramble_registration_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ramble_registration_groups_id_seq OWNER TO balade;

--
-- Name: ramble_registration_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.ramble_registration_groups_id_seq OWNED BY public.ramble_registration_groups.id;


--
-- Name: ramble_registrations; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.ramble_registrations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text NOT NULL,
    phone text,
    status text DEFAULT 'pending'::text NOT NULL,
    registration_date timestamp with time zone NOT NULL,
    confirmation_date timestamp with time zone,
    confirmation_deadline timestamp with time zone,
    cancellation_date timestamp with time zone,
    cancellation_reason text,
    ramble_id bigint NOT NULL,
    user_id bigint,
    group_id bigint
);


ALTER TABLE public.ramble_registrations OWNER TO balade;

--
-- Name: ramble_registrations_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.ramble_registrations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ramble_registrations_id_seq OWNER TO balade;

--
-- Name: ramble_registrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.ramble_registrations_id_seq OWNED BY public.ramble_registrations.id;


--
-- Name: rambles; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.rambles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text NOT NULL,
    description text,
    type text DEFAULT 'Découverte générale'::text NOT NULL,
    date timestamp with time zone,
    location text,
    meeting_point text,
    max_participants bigint,
    difficulty text DEFAULT 'Facile'::text NOT NULL,
    estimated_duration text,
    equipment_needed text,
    prerequisites text,
    cover_image text,
    additional_documents_url text,
    is_cancelled boolean DEFAULT false NOT NULL,
    cancellation_date timestamp with time zone,
    cancellation_reason text,
    payment_guide_id bigint,
    payment_enabled boolean DEFAULT false NOT NULL,
    payment_required boolean DEFAULT false NOT NULL
);


ALTER TABLE public.rambles OWNER TO balade;

--
-- Name: rambles_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.rambles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.rambles_id_seq OWNER TO balade;

--
-- Name: rambles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.rambles_id_seq OWNED BY public.rambles.id;


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.role_permissions (
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL
);


ALTER TABLE public.role_permissions OWNER TO balade;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);


ALTER TABLE public.roles OWNER TO balade;

--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.roles_id_seq OWNER TO balade;

--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: seeds; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.seeds (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);


ALTER TABLE public.seeds OWNER TO balade;

--
-- Name: seeds_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.seeds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.seeds_id_seq OWNER TO balade;

--
-- Name: seeds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.seeds_id_seq OWNED BY public.seeds.id;


--
-- Name: user_permission_overrides; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.user_permission_overrides (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    is_granted boolean NOT NULL
);


ALTER TABLE public.user_permission_overrides OWNER TO balade;

--
-- Name: user_permission_overrides_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.user_permission_overrides_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_permission_overrides_id_seq OWNER TO balade;

--
-- Name: user_permission_overrides_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.user_permission_overrides_id_seq OWNED BY public.user_permission_overrides.id;


--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.user_profiles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    first_name text NOT NULL,
    last_name text,
    avatar_name text,
    phone text,
    address_id bigint,
    user_id bigint NOT NULL
);


ALTER TABLE public.user_profiles OWNER TO balade;

--
-- Name: user_profiles_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.user_profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_profiles_id_seq OWNER TO balade;

--
-- Name: user_profiles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.user_profiles_id_seq OWNED BY public.user_profiles.id;


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.user_roles (
    user_id bigint NOT NULL,
    role_id bigint NOT NULL
);


ALTER TABLE public.user_roles OWNER TO balade;

--
-- Name: users; Type: TABLE; Schema: public; Owner: balade
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text NOT NULL,
    authentication_code text,
    authentication_expire_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO balade;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: balade
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO balade;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: balade
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: addresses id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.addresses ALTER COLUMN id SET DEFAULT nextval('public.addresses_id_seq'::regclass);


--
-- Name: guides id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.guides ALTER COLUMN id SET DEFAULT nextval('public.guides_id_seq'::regclass);


--
-- Name: payments id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.payments ALTER COLUMN id SET DEFAULT nextval('public.payments_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: ramble_guides id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_guides ALTER COLUMN id SET DEFAULT nextval('public.ramble_guides_id_seq'::regclass);


--
-- Name: ramble_prices id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_prices ALTER COLUMN id SET DEFAULT nextval('public.ramble_prices_id_seq'::regclass);


--
-- Name: ramble_registration_groups id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registration_groups ALTER COLUMN id SET DEFAULT nextval('public.ramble_registration_groups_id_seq'::regclass);


--
-- Name: ramble_registrations id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registrations ALTER COLUMN id SET DEFAULT nextval('public.ramble_registrations_id_seq'::regclass);


--
-- Name: rambles id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.rambles ALTER COLUMN id SET DEFAULT nextval('public.rambles_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: seeds id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.seeds ALTER COLUMN id SET DEFAULT nextval('public.seeds_id_seq'::regclass);


--
-- Name: user_permission_overrides id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_permission_overrides ALTER COLUMN id SET DEFAULT nextval('public.user_permission_overrides_id_seq'::regclass);


--
-- Name: user_profiles id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_profiles ALTER COLUMN id SET DEFAULT nextval('public.user_profiles_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: addresses; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.addresses (id, created_at, updated_at, deleted_at, number, route, optional_route, city, zip_code, country, latitude, longitude) FROM stdin;
\.


--
-- Data for Name: guides; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.guides (id, created_at, updated_at, deleted_at, first_name, last_name, email, phone, bio, experience, specialties, languages, certification_level, avatar, is_active, emergency_contact_name, emergency_contact_phone, stripe_account_id, stripe_public_key, stripe_secret_key, stripe_webhook_secret, payment_enabled, user_id) FROM stdin;
2	2025-09-11 14:47:32.364735+00	2026-05-25 15:31:44.0309+00	\N	Vera	Lorenzetti	contact@baladeecologique.com	PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	Je m’appelle Vera Lorenzetti, je suis diplômée d’une licence en Sciences de l’environnement de l’université de Lausanne et d’un master en Écologie de l’université de Bâle, où j’ai aussi complété mes études par un diplôme mineur en Développement durable. D’origine suisse, j’ai un parcours recherche pluridisciplinaire qui m’a permis de me spécialiser dans des enjeux écologiques actuels tels le changement climatique et les menaces sur la biodiversité. J’ai étudié la botanique et je me perfectionne dans la mycologie grâce à la Société Mycologique de Rennes, dont je suis la secrétaire depuis 2021. Je suis formatrice auprès de deux BTS sur le campus The Land.\n\nDepuis mon arrivée à Rennes en 2020, j’ai travaillé dans des associations d’éducation à l’environnement et de vulgarisation scientifique, en apprenant à simplifier et transmettre à tout type de public mes connaissances naturalistes. Cela m’a permis d’y voir plus clair dans mon parcours académique et de trouver le dénominateur commun qui m’anime dans mes recherches et ma transmission : revenir aux fondamentaux de la Vie pour construire un système durable et résilient dans lequel chacun puisse trouver sa place et s’épanouir. C’est dans ce but ultime que l’EcoLogique est née 🙂\n\nJ’adore observer et découvrir ce qui nous entoure quand on est dans la nature, surtout tout ce qui est petit et bizarre…c’est pourquoi je continue à me former et à m’intéresser à de nouveaux domaines !	Après mon master en écologie en 2019, j'ai travaillé comme animatrice scientifique à l'Espace des Sciences à Rennes, puis j'ai monté mon propre projet : Balade EcoLogique. Je suis guide nature depuis 2021. J'organise aussi des weekends pour le Club Chilowé et je suis la secrétaire de la Société Mycologique de Rennes.	Mycologie, botanique, écologie	Français, italien, anglais	professionnel	2.jpg	t		PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	\N	pk_live_51SGFGSFXETK94KjjEfu0rGIZbyTfFGXW5DDE4z1uoFsXganGUMdaBCCZu7s2xpbC7XnVoZPAdv2PutIitltKTpTY0049b5kgsF	YDo+aBMgzOaR8Bfn5hfdRxD/vAi2COAJmSA6wvAHxsRNsjxDPRzRRSGvzgt9UsXfUY6NV3ospk6AFs9gzIK26yUKmBKQKM6/fSqZ7EhxCVBltS+DRuM8LhvkhN0NWJmnpe/D7VXLNvTOjX4UTLh19iRg1BTTl4ubSg7dOxL4QGKyd8LYZyMD	vPp+FNKVRhRJBPkeZTAJCsZej7R1GlwI8QfcLminOEOoXkG7bWAu1Xdz1A/ew/rK398GzFwzXrS/EMx3izj11iDn	t	\N
1	2025-09-07 09:50:48.71232+00	2025-10-11 17:35:37.890184+00	\N	Isabelle	Cheval	isabellecheval8@gmail.com	PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	Bonjour ami(e)s de la Nature ☘️🐝🪶🥀🐦🌲\nJe suis Isabelle Cheval, passionnée par tout le Vivant et partenaire de Balade EcoLogique.\n🍀🫖 Je confectionne des tisanes personnalisées à la demande.\n🥀🐝 Je propose au fil des saisons et tout au long de l’année des ateliers en lien avec la Nature.\n🌲🐦Vous y trouverez des balades de reconnaissance des plantes sauvages comestibles de Saison (jeunes pousses printanières, arbres comestibles, fleurs à manger, racines à découvrir,…), une formule de cuisine sauvage à la journée, des ateliers de confection de baumes de plantes, de sirops, de vins médicinaux, de tisanes,…. \n🍀🍀J’anime aussi des sorties ou ateliers en partenariat avec celles et ceux qui sont passionnés par la Nature et le bien-être ! \n🌲🐦Pour découvrir le programme que je transmets au fil des Saisons, vous pouvez me le demander par ici : isabellecheval8@gmail.com\n🪶🐝Et vous pouvez aussi me retrouver sur LinkedIn.\nÀ bientôt !	Paysanne en plantes aromatiques et cueillette sauvage, installée depuis 2020 sur la Micro-Ferme Les Vies La Joie, route de la Guinois à Pléchâtel.	Ceuillette et cuisine des plantes sauvages.	Français	professionnel	1.jpg	t		PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	\N	\N	\N	\N	f	\N
5	2025-09-26 08:22:55.992428+00	2025-09-26 08:25:55.493374+00	\N	Valentin	Legendre	valentinvllegendre@gmail.com	PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	Salut, moi c’est Valentin Legendre, j’ai en poche un BTS en Gestion et Protection de la Nature couplé\navec une Licence en Biologie des Organismes, Écologie Éthologie et Évolution. Depuis que j’ai quitté l’Île de la Réunion qui m’a vu grandir, j’ai vadrouillé entre la Normandie, le Finistère, le Pays Basque et finalement Rennes où je vis actuellement.\n\nAu cours de mes études, j’ai découvert la nécessité de joindre les deux composantes complémentaires que sont la gestion et l’animation. Le constat est logique : pour protéger la nature, il faut d’abord la connaître. Cette phrase prend tout son sens pour moi quand l’on se rend compte que c’est en apprenant à connaître la nature que l’on apprend à l’apprécier, et ce faisant, développer une attention envers elle. Par là on peut comprendre l’impact de nos actes ainsi que notre façon de voir le vivant, de se voir au sein du vivant aussi.\n\nMon envie en tant qu’animateur est de partager mes connaissances et mes expériences afin de générer un intérêt, une volonté de comprendre la nature et in fine, de vouloir la protéger à son tour.\nAnimateur en colo depuis presque une dizaine d’étés auprès de jeunes enfants comme de jeunes adultes, et animateur périscolaire le reste de l’année ; la nature est quelque chose que je souhaite faire découvrir dès le plus jeune âge. Afin que le cheminement de pensée, et les comportements qui en découleront, puissent être orientés par cet intérêt pour la nature et le vivant si important à mes yeux.\nMême si mon champ de prédilection reste l’ornithologie, je possède de solides bases concernant toutes les petites (et moins petites) bêtes qui courent, marchent, nagent, rampent, volent ou sautent. J’ai à cœur de continuer d’apprendre car plus qu’un moteur, échanger avec autrui dans un but de découverte mutuelle est pour moi une part importante du travail de sensibilisation et de compréhension du vivant.		Oiseaux, amphibiens, faune et biodiversité		professionnel	5.jpg	t		PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	\N	\N	\N	\N	f	\N
6	2026-04-28 14:40:50.405145+00	2026-05-25 14:34:32.488522+00	\N	Val	Fortina	valfortinapro@proton.me	PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )			Botanique, usages de la nature, histoire de Brocéliande	Français, anglais	professionnel	6.jpg	t		PhoneNumber(isoCode: IsoCode.FR, countryCode: 33, nsn: )	\N	\N	\N	\N	f	115
\.


--
-- Data for Name: payments; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.payments (id, created_at, updated_at, deleted_at, stripe_payment_intent_id, stripe_charge_id, amount, currency, status, payment_method, failure_reason, registration_id, group_id, payer_email, payer_name, guide_id, paid_at, refunded_at, refund_amount) FROM stdin;
22	2025-10-24 16:48:48.570271+00	2025-10-24 16:48:54.899651+00	\N	pi_3SLoAWFXETK94Kjj0sIjsLGJ	\N	1500	eur	succeeded	card	\N	95	\N	davoctau@gmail.com	David  Octau 	2	2025-10-24 16:48:54.897419+00	\N	\N
1	2025-10-11 17:49:42.571418+00	2025-10-11 17:49:47.037852+00	\N	pi_3SH6vKFXETK94Kjj1L4H3rX9	\N	1500	eur	succeeded	card	\N	81	\N	victordenis01@gmail.com	Victor DENIS	2	2025-10-11 17:49:47.03565+00	\N	\N
2	2025-10-17 05:11:36.395357+00	2025-10-17 05:11:36.395357+00	\N	pi_3SJ5wyFXETK94Kjj27dJyK4U	\N	1500	eur	pending	card	\N	67	\N	r.delannay@orange.fr	Renaud Delannay	2	\N	\N	\N
3	2025-10-17 05:13:08.494537+00	2025-10-17 05:13:11.536223+00	\N	pi_3SJ5ySFXETK94Kjj24rhekbX	\N	1500	eur	succeeded	card	\N	67	\N	r.delannay@orange.fr	Renaud Delannay	2	2025-10-17 05:13:11.532828+00	\N	\N
4	2025-10-17 06:13:43.583235+00	2025-10-17 06:13:45.404087+00	\N	pi_3SJ6v5FXETK94Kjj1pO24hsP	\N	1500	eur	succeeded	card	\N	73	\N	nadegecorbe3576@gmail.com	Nadège  Lécrivain	2	2025-10-17 06:13:45.402084+00	\N	\N
5	2025-10-17 08:04:52.401854+00	2025-10-17 08:04:54.116324+00	\N	pi_3SJ8eeFXETK94Kjj1VdhyQPt	\N	1500	eur	succeeded	card	\N	55	\N	pyfab.balcon@wanadoo.fr	fabienne balcon	2	2025-10-17 08:04:54.113896+00	\N	\N
6	2025-10-19 17:46:06.896363+00	2025-10-19 17:46:06.896363+00	\N	pi_3SK0gEFXETK94Kjj1ApwlPij	\N	1500	eur	pending	card	\N	86	\N	laura.giommi@live.fr	Laura  Giommi 	2	\N	\N	\N
7	2025-10-19 17:46:08.486987+00	2025-10-19 17:46:08.486987+00	\N	pi_3SK0gGFXETK94Kjj1nh3V3aG	\N	1500	eur	pending	card	\N	87	\N	solene.barbe@hotmail.fr	Solène  Barbé 	2	\N	\N	\N
8	2025-10-19 17:46:14.246496+00	2025-10-19 17:46:14.246496+00	\N	pi_3SK0gMFXETK94Kjj0rfqI3mz	\N	1500	eur	pending	card	\N	86	\N	laura.giommi@live.fr	Laura  Giommi 	2	\N	\N	\N
9	2025-10-19 17:46:31.118926+00	2025-10-19 17:46:31.118926+00	\N	pi_3SK0gdFXETK94Kjj13Eooz2g	\N	1500	eur	pending	card	\N	87	\N	solene.barbe@hotmail.fr	Solène  Barbé 	2	\N	\N	\N
10	2025-10-19 17:46:55.404728+00	2025-10-19 17:47:00.226508+00	\N	pi_3SK0h1FXETK94Kjj1JzZaRxT	\N	1500	eur	succeeded	card	\N	87	\N	solene.barbe@hotmail.fr	Solène  Barbé 	2	2025-10-19 17:47:00.224123+00	\N	\N
29	2025-10-29 17:41:34.385723+00	2025-10-29 17:41:40.045826+00	\N	pi_3SNdNKFXETK94Kjj1Fxra0JW	\N	1500	eur	succeeded	card	\N	123	\N	titouan.millon3@gmail.com	Titouan Millon	2	2025-10-29 17:41:40.043854+00	\N	\N
11	2025-10-19 17:47:49.051762+00	2025-10-19 17:47:55.311249+00	\N	pi_3SK0hsFXETK94Kjj2ht5ni6x	\N	1500	eur	succeeded	card	\N	86	\N	laura.giommi@live.fr	Laura  Giommi 	2	2025-10-19 17:47:55.309051+00	\N	\N
12	2025-10-19 17:58:05.532657+00	2025-10-19 17:58:05.532657+00	\N	pi_3SK0rpFXETK94Kjj12vc2sVB	\N	1500	eur	pending	card	\N	98	\N	demeulenaere.manon@yahoo.fr	manon demeulenaere	2	\N	\N	\N
23	2025-10-24 16:49:23.058862+00	2025-10-24 16:49:28.535591+00	\N	pi_3SLoB4FXETK94Kjj2ZmgbxCW	\N	1500	eur	succeeded	card	\N	47	\N	elodie.octau@gmail.com	Elodie Octau	2	2025-10-24 16:49:28.533479+00	\N	\N
13	2025-10-19 17:58:41.740085+00	2025-10-19 17:58:48.092746+00	\N	pi_3SK0sPFXETK94Kjj0QtJto6S	\N	1500	eur	succeeded	card	\N	98	\N	demeulenaere.manon@yahoo.fr	manon demeulenaere	2	2025-10-19 17:58:48.090771+00	\N	\N
14	2025-10-21 15:32:58.387964+00	2025-10-21 15:32:58.387964+00	\N	pi_3SKhYUFXETK94Kjj1HFeiR0o	\N	1500	eur	pending	card	\N	78	\N	iven.lelouedec@gmail.com	iven Le Louedec	2	\N	\N	\N
15	2025-10-21 15:33:43.879946+00	2025-10-21 15:33:46.235988+00	\N	pi_3SKhZDFXETK94Kjj05K6ppRU	\N	1500	eur	succeeded	card	\N	78	\N	iven.lelouedec@gmail.com	iven Le Louedec	2	2025-10-21 15:33:46.232884+00	\N	\N
16	2025-10-22 11:35:37.365556+00	2025-10-22 11:35:37.365556+00	\N	pi_3SL0KLFXETK94Kjj055Vsl96	\N	5500	eur	pending	card	\N	\N	23	daheronf@gmail.com	DAHERON 	2	\N	\N	\N
17	2025-10-22 11:36:47.227508+00	2025-10-22 11:37:28.206947+00	\N	pi_3SL0LTFXETK94Kjj09bWzQOl	\N	5500	eur	succeeded	card	\N	\N	23	daheronf@gmail.com	DAHERON 	2	2025-10-22 11:37:28.20491+00	\N	\N
18	2025-10-24 05:04:38.483481+00	2025-10-24 05:04:38.483481+00	\N	pi_3SLdB4FXETK94Kjj1iJvzajC	\N	1500	eur	pending	card	\N	13	\N	lolia.cozette@gmail.com	Lolita COZETTE 	2	\N	\N	\N
19	2025-10-24 05:04:44.65677+00	2025-10-24 05:04:44.65677+00	\N	pi_3SLdBAFXETK94Kjj062tJTkA	\N	1500	eur	pending	card	\N	13	\N	lolia.cozette@gmail.com	Lolita COZETTE 	2	\N	\N	\N
20	2025-10-24 05:04:56.765737+00	2025-10-24 05:05:04.336768+00	\N	pi_3SLdBMFXETK94Kjj2a8RmKIT	\N	1500	eur	succeeded	card	\N	13	\N	lolia.cozette@gmail.com	Lolita COZETTE 	2	2025-10-24 05:05:04.336304+00	\N	\N
21	2025-10-24 10:17:35.120125+00	2025-10-24 10:17:37.59245+00	\N	pi_3SLi3vFXETK94Kjj1oD6BgAe	\N	1500	eur	succeeded	card	\N	127	\N	pauline-guilbaud@hotmail.fr	Pauline Guilbaud 	2	2025-10-24 10:17:37.589075+00	\N	\N
30	2025-10-30 06:59:33.176078+00	2025-10-30 06:59:35.862327+00	\N	pi_3SNppZFXETK94Kjj0t4oGRHL	\N	1500	eur	succeeded	card	\N	12	\N	Vandepeutte.coline@hotmail.fr	coline Vandepeutte 	2	2025-10-30 06:59:35.859449+00	\N	\N
24	2025-10-26 18:46:27.066209+00	2025-10-26 18:46:31.772624+00	\N	pi_3SMYxSFXETK94Kjj0IRD2pDk	\N	1500	eur	succeeded	card	\N	135	\N	maeliss.monbon@lilo.org	Maëliss  Monbon	2	2025-10-26 18:46:31.770599+00	\N	\N
25	2025-10-28 05:46:40.388228+00	2025-10-28 05:46:40.388228+00	\N	pi_3SN5jwFXETK94Kjj2bLiyHOo	\N	1500	eur	pending	card	\N	71	\N	adline.leon@gmail.com	adeline Léon 	2	\N	\N	\N
26	2025-10-28 05:48:20.328171+00	2025-10-28 05:48:24.545455+00	\N	pi_3SN5lYFXETK94Kjj25SRB4TR	\N	1500	eur	succeeded	card	\N	71	\N	adline.leon@gmail.com	adeline Léon 	2	2025-10-28 05:48:24.543634+00	\N	\N
27	2025-10-29 07:02:43.128742+00	2025-10-29 07:02:43.128742+00	\N	pi_3SNTP5FXETK94Kjj1qCYFlr4	\N	3000	eur	pending	card	\N	\N	21	nathalie.strugalski@hotmail.fr	Delvincourt 	2	\N	\N	\N
31	2025-11-20 12:32:24.184272+00	2025-11-20 12:32:26.133543+00	\N	pi_3SVX2CFXETK94Kjj1FK9N6jA	\N	3000	eur	succeeded	card	\N	\N	30	emilie.gillier@lilo.org	Victor Gaudard	2	2025-11-20 12:32:26.130841+00	\N	\N
28	2025-10-29 07:03:06.372482+00	2025-10-29 07:03:12.518345+00	\N	pi_3SNTPSFXETK94Kjj0i8slH3x	\N	3000	eur	succeeded	card	\N	\N	21	nathalie.strugalski@hotmail.fr	Delvincourt 	2	2025-10-29 07:03:12.516291+00	\N	\N
35	2026-04-22 17:01:12.635594+00	2026-04-22 17:01:12.635594+00	\N	pi_3TP49EFXETK94Kjj0PmpKxcN	\N	1000	eur	pending	card	\N	156	\N	vielderennes@orange.fr	anne viel	2	\N	\N	\N
32	2025-11-26 05:49:31.948168+00	2025-11-26 05:49:31.948168+00	\N	pi_3SXbbbFXETK94Kjj2PJgxsr6	\N	2000	eur	pending	card	\N	\N	29	vincent.virginie@gmail.com	Vincent Virginie 	2	\N	\N	\N
33	2025-11-26 05:49:47.173232+00	2025-11-26 05:49:48.937279+00	\N	pi_3SXbbrFXETK94Kjj1JNkc0Ls	\N	2000	eur	succeeded	card	\N	\N	29	vincent.virginie@gmail.com	Vincent Virginie 	2	2025-11-26 05:49:48.934073+00	\N	\N
34	2026-04-22 16:59:20.267218+00	2026-04-22 16:59:22.358181+00	\N	pi_3TP47QFXETK94Kjj2TOVZqm2	\N	1000	eur	succeeded	card	\N	157	\N	vieldelangon@wanadoo.fr	Alain  viel 	2	2026-04-22 16:59:22.355923+00	\N	\N
36	2026-04-22 17:01:42.16457+00	2026-04-22 17:01:50.263759+00	\N	pi_3TP49iFXETK94Kjj0YvUCyJy	\N	1000	eur	succeeded	card	\N	156	\N	vielderennes@orange.fr	anne viel	2	2026-04-22 17:01:50.261856+00	\N	\N
37	2026-04-22 19:56:32.341779+00	2026-04-22 19:56:35.308377+00	\N	pi_3TP6suFXETK94Kjj13sOtW4s	\N	2400	eur	succeeded	card	\N	\N	32	maeva.cadeau@orange.fr	CADEAU Maéva 	2	2026-04-22 19:56:35.306041+00	\N	\N
39	2026-05-01 08:55:56.750217+00	2026-05-01 08:55:56.750217+00	\N	pi_3TSCrYFXETK94Kjj2ze0TgHU	\N	4000	eur	pending	card	\N	158	\N	vielderennes@orange.fr	anne viel	2	\N	\N	\N
38	2026-05-01 08:53:48.713636+00	2026-05-01 08:54:45.020029+00	\N	pi_3TSCpUFXETK94Kjj0klduEwr	\N	4000	eur	succeeded	card	\N	159	\N	vieldelangon@wanadoo.fr	alain viel	2	2026-05-01 08:54:45.018097+00	\N	\N
40	2026-05-01 08:56:28.832036+00	2026-05-01 08:56:32.975578+00	\N	pi_3TSCs4FXETK94Kjj2h5KLCDd	\N	4000	eur	succeeded	card	\N	158	\N	vielderennes@orange.fr	anne viel	2	2026-05-01 08:56:32.973449+00	\N	\N
41	2026-05-02 04:44:34.893546+00	2026-05-02 04:44:34.893546+00	\N	pi_3TSVPqFXETK94Kjj2XNgRCou	\N	4000	eur	pending	card	\N	167	\N	fourgautdorine@yahoo.fr	Dorine  Fourgaut 	2	\N	\N	\N
42	2026-05-02 04:45:23.758367+00	2026-05-02 04:45:27.987989+00	\N	pi_3TSVQdFXETK94Kjj2aETOSC7	\N	4000	eur	succeeded	card	\N	167	\N	fourgautdorine@yahoo.fr	Dorine  Fourgaut 	2	2026-05-02 04:45:27.98611+00	\N	\N
43	2026-06-03 07:05:52.196071+00	2026-06-03 07:05:52.196071+00	\N	pi_3Te8s7FXETK94Kjj2wh8AQyE	\N	2000	eur	pending	card	\N	174	\N	agathesene0@gmail.com	Agathe Séné	2	\N	\N	\N
44	2026-06-03 07:05:54.486166+00	2026-06-03 07:05:54.486166+00	\N	pi_3Te8sAFXETK94Kjj2FqH6dpK	\N	2000	eur	pending	card	\N	174	\N	agathesene0@gmail.com	Agathe Séné	2	\N	\N	\N
45	2026-06-03 07:07:51.869446+00	2026-06-03 07:07:55.046087+00	\N	pi_3Te8u3FXETK94Kjj0Ox49uPl	\N	2000	eur	succeeded	card	\N	174	\N	agathesene0@gmail.com	Agathe Séné	2	2026-06-03 07:07:55.040497+00	\N	\N
46	2026-06-03 07:08:59.826859+00	2026-06-03 07:08:59.826859+00	\N	pi_3Te8v9FXETK94Kjj221OUj72	\N	2000	eur	pending	card	\N	173	\N	agathe.senedu56@gmail.com	Anaëlle Morel	2	\N	\N	\N
47	2026-06-03 07:09:02.165054+00	2026-06-03 07:09:02.165054+00	\N	pi_3Te8vBFXETK94Kjj0a2EQdG7	\N	2000	eur	pending	card	\N	173	\N	agathe.senedu56@gmail.com	Anaëlle Morel	2	\N	\N	\N
48	2026-06-03 07:09:30.405609+00	2026-06-03 07:09:35.077821+00	\N	pi_3Te8vdFXETK94Kjj1NPStuwr	\N	2000	eur	succeeded	card	\N	173	\N	agathe.senedu56@gmail.com	Anaëlle Morel	2	2026-06-03 07:09:35.073909+00	\N	\N
49	2026-06-05 10:29:57.721904+00	2026-06-05 10:29:57.721904+00	\N	pi_3Tev0jFXETK94Kjj16fzIY7t	\N	4000	eur	pending	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
50	2026-06-05 10:30:35.768941+00	2026-06-05 10:30:35.768941+00	\N	pi_3Tev1LFXETK94Kjj1ZMCQvtl	\N	4000	eur	pending	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
51	2026-06-05 10:30:55.043788+00	2026-06-05 10:30:55.043788+00	\N	pi_3Tev1eFXETK94Kjj1lwNLzMw	\N	4000	eur	pending	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
52	2026-06-05 10:31:27.15417+00	2026-06-05 10:31:27.15417+00	\N	pi_3Tev2AFXETK94Kjj1lFEuL3i	\N	4000	eur	pending	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
53	2026-06-05 10:32:07.947211+00	2026-06-05 10:32:07.947211+00	\N	pi_3Tev2pFXETK94Kjj1xkt8zVK	\N	4000	eur	pending	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
54	2026-06-05 10:32:24.928246+00	2026-06-05 10:32:24.928246+00	\N	pi_3Tev36FXETK94Kjj2zjZnTrE	\N	4000	eur	pending	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
55	2026-06-05 10:32:38.212492+00	2026-06-05 10:32:41.372752+00	\N	pi_3Tev3KFXETK94Kjj1Aa4mYUa	\N	4000	eur	succeeded	card	\N	178	\N	emibouu@gmail.com	Emilie Bourgeois	2	2026-06-05 10:32:41.36837+00	\N	\N
56	2026-06-26 10:44:50.782303+00	2026-06-26 10:44:50.782303+00	\N	pi_3TmXFeFXETK94Kjj2C1enrd5	\N	4000	eur	pending	card	\N	180	\N	amelaint@yahoo.fr	Virginie  Chérel	2	\N	\N	\N
57	2026-06-26 10:44:54.484194+00	2026-06-26 10:44:54.484194+00	\N	pi_3TmXFiFXETK94Kjj2ESi6pbC	\N	4000	eur	pending	card	\N	180	\N	amelaint@yahoo.fr	Virginie  Chérel	2	\N	\N	\N
58	2026-06-26 10:46:22.725792+00	2026-06-26 10:46:26.085017+00	\N	pi_3TmXH8FXETK94Kjj08xbtJop	\N	4000	eur	succeeded	card	\N	180	\N	amelaint@yahoo.fr	Virginie  Chérel	2	2026-06-26 10:46:26.079925+00	\N	\N
59	2026-06-27 03:35:34.391132+00	2026-06-27 03:35:36.881684+00	\N	pi_3Tmn1mFXETK94Kjj1lPt6zpc	\N	2000	eur	succeeded	card	\N	196	\N	2.yassine.amar@gmail.com	Yassine Amar	2	2026-06-27 03:35:36.877384+00	\N	\N
60	2026-07-01 07:26:48.929153+00	2026-07-01 07:26:48.929153+00	\N	pi_3ToIXkFXETK94Kjj1HHkMIIT	\N	2500	eur	pending	card	\N	189	\N	lea.rion310@gmail.com	Léa RION	2	\N	\N	\N
61	2026-07-01 07:27:18.181784+00	2026-07-01 07:27:21.424726+00	\N	pi_3ToIYEFXETK94Kjj0uJlRYVO	\N	2500	eur	succeeded	card	\N	189	\N	lea.rion310@gmail.com	Léa RION	2	2026-07-01 07:27:21.419603+00	\N	\N
62	2026-07-02 17:20:42.119596+00	2026-07-02 17:20:42.119596+00	\N	pi_3TooI1FXETK94Kjj1Jefih6r	\N	10000	eur	pending	card	\N	\N	36	vielderennes@orange.fr	Viel Alain et anne	2	\N	\N	\N
63	2026-07-02 17:21:49.23478+00	2026-07-02 17:21:52.988877+00	\N	pi_3TooJ7FXETK94Kjj2SfGoG7c	\N	10000	eur	succeeded	card	\N	\N	36	vielderennes@orange.fr	Viel Alain 	2	2026-07-02 17:21:52.98232+00	\N	\N
64	2026-07-02 19:59:22.201522+00	2026-07-02 19:59:26.014504+00	\N	pi_3ToqlaFXETK94Kjj07RqtyG6	\N	2000	eur	succeeded	card	\N	181	\N	jtixier@free.fr	Jacques  tixier	2	2026-07-02 19:59:26.009331+00	\N	\N
65	2026-07-03 11:04:38.131478+00	2026-07-03 11:04:38.131478+00	\N	pi_3Tp4tdFXETK94Kjj12Cz6yPq	\N	5000	eur	pending	card	\N	197	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
66	2026-07-03 11:04:50.302521+00	2026-07-03 11:04:50.302521+00	\N	pi_3Tp4tqFXETK94Kjj06PfJNcM	\N	5000	eur	pending	card	\N	197	\N	emibouu@gmail.com	Emilie Bourgeois	2	\N	\N	\N
67	2026-07-03 11:06:09.166716+00	2026-07-03 11:06:12.253845+00	\N	pi_3Tp4v7FXETK94Kjj1rnrxNjF	\N	5000	eur	succeeded	card	\N	197	\N	emibouu@gmail.com	Emilie Bourgeois	2	2026-07-03 11:06:12.249978+00	\N	\N
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.permissions (id, created_at, updated_at, deleted_at, name, readable_name, description, category) FROM stdin;
1	2025-09-06 17:09:51.254667+00	2025-09-06 17:09:51.254667+00	\N	create:user	Créer un utilisateur	Permet de créer un nouvel utilisateur	\N
2	2025-09-06 17:09:51.255369+00	2025-09-06 17:09:51.255369+00	\N	read:user	Lire un utilisateur	Permet de lire les informations d'un utilisateur	\N
4	2025-09-06 17:09:51.256061+00	2025-09-06 17:09:51.256061+00	\N	update:user	Mettre à jour un utilisateur	Permet de mettre à jour les informations d'un utilisateur	\N
6	2025-09-06 17:09:51.256564+00	2025-09-06 17:09:51.256564+00	\N	delete:user	Supprimer un utilisateur	Permet de supprimer un utilisateur	\N
7	2025-09-06 17:09:51.256837+00	2025-09-06 17:09:51.256837+00	\N	delete:user:self	Supprimer son propre utilisateur	Permet de supprimer son propre compte	\N
8	2025-09-06 17:09:51.26019+00	2025-09-06 17:09:51.26019+00	\N	create:ramble	Créer une balade	Permet de créer une nouvelle balade	\N
11	2025-09-06 17:09:51.261018+00	2025-09-06 17:09:51.261018+00	\N	update:ramble	Mettre à jour une balade	Permet de mettre à jour les informations d'une balade	\N
13	2025-09-06 17:09:51.261461+00	2025-09-06 17:09:51.261461+00	\N	delete:ramble	Supprimer une balade	Permet de supprimer une balade	\N
14	2025-09-06 17:09:51.261695+00	2025-09-06 17:09:51.261695+00	\N	delete:ramble:self	Supprimer sa propre balade	Permet de supprimer ses propres informations	\N
15	2025-10-05 11:16:15.629274+00	2025-10-05 11:16:15.629274+00	\N	cancel:ramble	Annuler une balade	Permet d'annuler une balade avec un motif	\N
16	2025-10-05 11:16:15.632364+00	2025-10-05 11:16:15.632364+00	\N	manage:registration	Gérer les inscriptions	Permet de gérer toutes les inscriptions (lecture, modification, suppression)	\N
17	2025-10-05 11:16:15.632618+00	2025-10-05 11:16:15.632618+00	\N	view:all-registrations	Voir toutes les inscriptions	Permet de voir toutes les inscriptions de toutes les balades	\N
18	2025-10-05 11:16:15.632894+00	2025-10-05 11:16:15.632894+00	\N	update:registration-status	Modifier le statut des inscriptions	Permet de modifier le statut des inscriptions (confirmer, annuler, etc.)	\N
19	2025-10-05 11:16:15.633178+00	2025-10-05 11:16:15.633178+00	\N	update:registration-details	Modifier les détails d'inscription	Permet de modifier les informations personnelles des inscriptions	\N
20	2025-10-05 11:16:15.633454+00	2025-10-05 11:16:15.633454+00	\N	bulk:registration-actions	Actions en lot sur les inscriptions	Permet d'effectuer des actions en lot sur plusieurs inscriptions	\N
21	2025-10-11 17:20:19.232254+00	2025-10-11 17:20:19.232254+00	\N	manage:payments	\N	Full payment management - create, view, refund payments	\N
22	2025-10-11 17:20:19.232654+00	2025-10-11 17:20:19.232654+00	\N	view:payments	\N	View payment information	\N
23	2025-10-11 17:20:19.232923+00	2025-10-11 17:20:19.232923+00	\N	configure:guide-payments	\N	Configure guide payment settings (Stripe credentials)	\N
24	2025-10-11 17:20:19.233188+00	2025-10-11 17:20:19.233188+00	\N	refund:payments	\N	Process payment refunds	\N
25	2025-10-11 17:20:19.233429+00	2025-10-11 17:20:19.233429+00	\N	webhook:payments	\N	Handle payment webhook events	\N
26	2026-05-25 14:08:06.629214+00	2026-05-25 14:08:06.629214+00	\N	configure:guide-payments:self	Configurer ses paiements guide	Configure own guide payment settings (Stripe credentials)	\N
27	2026-05-25 14:08:06.632079+00	2026-05-25 14:08:06.632079+00	\N	view:registrations:self	Voir les inscriptions de ses balades	View registrations for own guided rambles	\N
28	2026-05-25 14:08:06.632356+00	2026-05-25 14:08:06.632356+00	\N	manage:registration:self	Gérer les inscriptions de ses balades	Delete registrations for own guided rambles	\N
29	2026-05-25 14:08:06.632609+00	2026-05-25 14:08:06.632609+00	\N	update:registration-status:self	Modifier le statut des inscriptions de ses balades	Update registration status for own guided rambles	\N
30	2026-05-25 14:08:06.632861+00	2026-05-25 14:08:06.632861+00	\N	update:registration-details:self	Modifier les détails des inscriptions de ses balades	Update registration details for own guided rambles	\N
31	2026-05-25 14:08:06.633155+00	2026-05-25 14:08:06.633155+00	\N	bulk:registration-actions:self	Actions en lot sur les inscriptions de ses balades	Bulk actions on registrations for own guided rambles	\N
3	2025-09-06 17:09:51.255786+00	2026-05-25 14:34:32.493877+00	\N	read:user:self	Lire son propre utilisateur	Permet de lire ses propres informations	\N
5	2025-09-06 17:09:51.256296+00	2026-05-25 14:34:32.493877+00	\N	update:user:self	Mettre à jour son propre utilisateur	Permet de mettre à jour ses propres informations	\N
9	2025-09-06 17:09:51.260496+00	2026-05-25 14:34:32.493877+00	\N	read:ramble	Lire une balade	Permet de lire les informations d'une balade	\N
10	2025-09-06 17:09:51.26074+00	2026-05-25 14:34:32.493877+00	\N	read:ramble:self	Lire sa propre balade	Permet de lire ses propres informations	\N
12	2025-09-06 17:09:51.261247+00	2026-05-25 14:34:32.493877+00	\N	update:ramble:self	Mettre à jour sa propre balade	Permet de mettre à jour ses propres informations	\N
\.


--
-- Data for Name: ramble_guides; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.ramble_guides (id, created_at, updated_at, deleted_at, ramble_id, guide_id) FROM stdin;
124	\N	\N	\N	8	2
125	\N	\N	\N	17	5
126	\N	\N	\N	9	2
196	\N	\N	\N	30	2
130	\N	\N	\N	18	2
131	\N	\N	\N	12	2
133	\N	\N	\N	10	2
134	\N	\N	\N	11	2
138	\N	\N	\N	21	2
144	\N	\N	\N	23	2
24	\N	\N	\N	1	2
148	\N	\N	\N	24	2
27	\N	\N	\N	2	2
150	\N	\N	\N	25	5
155	\N	\N	\N	13	2
156	\N	\N	\N	13	1
157	\N	\N	\N	26	5
159	\N	\N	\N	22	2
226	\N	\N	\N	32	2
227	\N	\N	\N	32	6
228	\N	\N	\N	33	2
229	\N	\N	\N	34	2
168	\N	\N	\N	27	2
169	\N	\N	\N	27	5
232	\N	\N	\N	35	2
172	\N	\N	\N	28	2
173	\N	\N	\N	28	5
233	\N	\N	\N	35	6
113	\N	\N	\N	14	5
114	\N	\N	\N	15	5
118	\N	\N	\N	19	5
119	\N	\N	\N	20	5
121	\N	\N	\N	5	2
122	\N	\N	\N	6	2
123	\N	\N	\N	7	2
185	\N	\N	\N	29	2
189	\N	\N	\N	31	5
253	\N	\N	\N	36	6
254	\N	\N	\N	36	2
258	\N	\N	\N	40	6
261	\N	\N	\N	37	5
264	\N	\N	\N	38	6
265	\N	\N	\N	39	6
272	\N	\N	\N	41	6
275	\N	\N	\N	42	6
\.


--
-- Data for Name: ramble_prices; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.ramble_prices (id, created_at, updated_at, deleted_at, label, amount, ramble_id) FROM stdin;
210	2025-10-30 09:18:48.406097+00	2025-10-30 09:18:48.406097+00	\N	Adulte	15	21
211	2025-10-30 09:18:48.406097+00	2025-10-30 09:18:48.406097+00	\N	Enfant (-14 ans)	10	21
275	2026-04-24 10:29:54.162107+00	2026-04-24 10:29:54.162107+00	\N	Adulte	40	30
276	2026-04-24 10:29:54.162107+00	2026-04-24 10:29:54.162107+00	\N	Enfant, étudiant	20	30
218	2025-10-30 09:30:22.096861+00	2025-10-30 09:30:22.096861+00	\N	Adulte	10	23
219	2025-10-30 09:30:22.096861+00	2025-10-30 09:30:22.096861+00	\N	Enfant	8	23
164	2025-10-07 19:44:46.450818+00	2025-10-07 19:44:46.450818+00	\N	Adulte	15	14
165	2025-10-07 19:44:46.450818+00	2025-10-07 19:44:46.450818+00	\N	Enfant (-12 ans)	10	14
166	2025-10-08 13:40:22.799446+00	2025-10-08 13:40:22.799446+00	\N	Adulte	15	15
167	2025-10-08 13:40:22.799446+00	2025-10-08 13:40:22.799446+00	\N	Enfant (-12 ans)	10	15
224	2025-10-30 09:32:49.620496+00	2025-10-30 09:32:49.620496+00	\N	Adulte	10	24
225	2025-10-30 09:32:49.620496+00	2025-10-30 09:32:49.620496+00	\N	Enfant (-14 and)	8	24
226	2025-11-02 12:40:45.281262+00	2025-11-02 12:40:45.281262+00	\N	Adulte	10	25
179	2025-10-11 17:55:01.352191+00	2025-10-11 17:55:01.352191+00	\N	Adulte	15	19
180	2025-10-11 17:55:01.352191+00	2025-10-11 17:55:01.352191+00	\N	Enfant (-12 ans)	10	19
181	2025-10-11 17:55:10.69899+00	2025-10-11 17:55:10.69899+00	\N	Adulte	15	20
182	2025-10-11 17:55:10.69899+00	2025-10-11 17:55:10.69899+00	\N	Enfant (-12 ans)	10	20
227	2025-11-02 12:40:45.281262+00	2025-11-02 12:40:45.281262+00	\N	Enfant (-12 ans)	8	25
185	2025-10-15 20:13:48.877706+00	2025-10-15 20:13:48.877706+00	\N	Adulte	15	5
186	2025-10-15 20:13:48.877706+00	2025-10-15 20:13:48.877706+00	\N	Enfant (-12 ans)	10	5
187	2025-10-15 20:14:35.161222+00	2025-10-15 20:14:35.161222+00	\N	Adulte	15	6
188	2025-10-15 20:14:35.161222+00	2025-10-15 20:14:35.161222+00	\N	Enfant (-12 ans)	10	6
189	2025-10-15 20:14:51.25307+00	2025-10-15 20:14:51.25307+00	\N	Adulte	15	7
190	2025-10-15 20:14:51.25307+00	2025-10-15 20:14:51.25307+00	\N	Enfant (-12 ans)	10	7
191	2025-10-15 20:15:54.671456+00	2025-10-15 20:15:54.671456+00	\N	Adulte	15	8
192	2025-10-15 20:15:54.671456+00	2025-10-15 20:15:54.671456+00	\N	Enfant (-12 ans)	10	8
193	2025-10-17 21:13:28.941009+00	2025-10-17 21:13:28.941009+00	\N	Adulte	10	17
194	2025-10-17 21:13:28.941009+00	2025-10-17 21:13:28.941009+00	\N	Enfant (-12 ans)	8	17
195	2025-10-22 10:09:38.748743+00	2025-10-22 10:09:38.748743+00	\N	Adulte	15	9
196	2025-10-22 10:09:38.748743+00	2025-10-22 10:09:38.748743+00	\N	Enfant (-12 ans)	10	9
43	2025-09-24 09:12:58.147835+00	2025-09-24 09:12:58.147835+00	\N	Adulte	15	1
44	2025-09-24 09:12:58.147835+00	2025-09-24 09:12:58.147835+00	\N	Enfant (-12 ans)	10	1
231	2025-11-06 15:22:29.050097+00	2025-11-06 15:22:29.050097+00	\N	Adulte	60	13
228	2025-11-02 12:50:52.066461+00	2025-11-02 12:50:52.066461+00	\N	Adulte	10	26
229	2025-11-02 12:50:52.066461+00	2025-11-02 12:50:52.066461+00	\N	Enfant (-12 ans)	8	26
49	2025-09-24 09:13:44.325053+00	2025-09-24 09:13:44.325053+00	\N	Adulte	15	2
50	2025-09-24 09:13:44.325053+00	2025-09-24 09:13:44.325053+00	\N	Enfant (-12 ans)	10	2
232	2025-11-22 17:35:55.405074+00	2025-11-22 17:35:55.405074+00	\N	Adulte	10	22
233	2025-11-22 17:35:55.405074+00	2025-11-22 17:35:55.405074+00	\N	Enfant (-14 ans)	8	22
236	2026-03-13 17:35:08.766077+00	2026-03-13 17:35:08.766077+00	\N	Adultes et enfants	0	27
238	2026-03-13 17:38:18.497238+00	2026-03-13 17:38:18.497238+00	\N	Adultes et enfants	0	28
197	2025-10-22 10:09:55.060699+00	2025-10-22 10:09:55.060699+00	\N	Adulte	10	18
198	2025-10-22 10:09:55.060699+00	2025-10-22 10:09:55.060699+00	\N	Enfant (-12 ans)	8	18
200	2025-10-27 08:16:01.38302+00	2025-10-27 08:16:01.38302+00	\N	Adulte	15	12
201	2025-10-27 08:16:01.38302+00	2025-10-27 08:16:01.38302+00	\N	Enfant (-12 ans)	10	12
204	2025-10-27 08:16:52.762908+00	2025-10-27 08:16:52.762908+00	\N	Adulte	15	10
205	2025-10-27 08:16:52.762908+00	2025-10-27 08:16:52.762908+00	\N	Enfant (-12 ans)	10	10
202	2025-10-27 08:16:18.404963+00	2025-10-27 08:16:18.404963+00	\N	Adulte	15	11
203	2025-10-27 08:16:18.404963+00	2025-10-27 08:16:18.404963+00	\N	Enfant (-12 ans)	10	11
382	2026-06-15 13:46:49.760894+00	2026-06-15 13:46:49.760894+00	\N	Adultes	20	38
383	2026-06-15 13:46:49.760894+00	2026-06-15 13:46:49.760894+00	\N	Enfants -16 ans, RSA, étudiants	15	38
299	2026-05-06 10:29:36.953582+00	2026-05-06 10:29:36.953582+00	\N	Adulte	50	32
300	2026-05-06 10:29:36.953582+00	2026-05-06 10:29:36.953582+00	\N	Enfants -12 ans, étudiants, RSA	25	32
311	2026-05-20 09:49:30.409036+00	2026-05-20 09:49:30.409036+00	\N	Adulte	40	33
257	2026-04-06 10:28:12.63344+00	2026-04-06 10:28:12.63344+00	\N	Adulte	10	29
258	2026-04-06 10:28:12.63344+00	2026-04-06 10:28:12.63344+00	\N	Enfant	7	29
312	2026-05-20 09:49:30.409036+00	2026-05-20 09:49:30.409036+00	\N	Enfant -12 ans, étudiant, RSA	20	33
313	2026-05-20 10:23:15.546549+00	2026-05-20 10:23:15.546549+00	\N	Adulte	40	34
314	2026-05-20 10:23:15.546549+00	2026-05-20 10:23:15.546549+00	\N	Enfant -12 ans, étudiant, RSA	20	34
263	2026-04-22 12:54:20.255315+00	2026-04-22 12:54:20.255315+00	\N	Adulte	10	31
264	2026-04-22 12:54:20.255315+00	2026-04-22 12:54:20.255315+00	\N	Enfant (-12ans)	7	31
317	2026-05-25 14:51:19.642437+00	2026-05-25 14:51:19.642437+00	\N	Adulte	50	35
318	2026-05-25 14:51:19.642437+00	2026-05-25 14:51:19.642437+00	\N	Enfant -12 ans, étudiant, RSA	25	35
379	2026-06-15 13:43:52.26346+00	2026-06-15 13:43:52.26346+00	\N	Adulte	20	39
380	2026-06-15 13:43:52.26346+00	2026-06-15 13:43:52.26346+00	\N	Enfants -16 ans, RSA, étudiants	15	39
381	2026-06-15 13:43:52.26346+00	2026-06-15 13:43:52.26346+00	\N	Enfants -6 ans	0	39
391	2026-07-07 08:05:07.973287+00	2026-07-07 08:05:07.973287+00	\N	Adulte	20	41
392	2026-07-07 08:05:07.973287+00	2026-07-07 08:05:07.973287+00	\N	Enfants -16 ans, RSA, étudiants	15	41
394	2026-07-07 08:42:38.150464+00	2026-07-07 08:42:38.150464+00	\N	Adulte	20	42
357	2026-06-12 15:40:56.095573+00	2026-06-12 15:40:56.095573+00	\N	Adulte	50	36
358	2026-06-12 15:40:56.095573+00	2026-06-12 15:40:56.095573+00	\N	Enfant -12 ans, étudiant, RSA	25	36
368	2026-06-15 12:46:40.153907+00	2026-06-15 12:46:40.153907+00	\N	Adulte	20	40
369	2026-06-15 12:46:40.153907+00	2026-06-15 12:46:40.153907+00	\N	Enfants -16 ans, RSA, étudiants	15	40
375	2026-06-15 13:00:43.825153+00	2026-06-15 13:00:43.825153+00	\N	Adulte	10	37
376	2026-06-15 13:00:43.825153+00	2026-06-15 13:00:43.825153+00	\N	Enfant (- de 12 ans)	7	37
\.


--
-- Data for Name: ramble_registration_groups; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.ramble_registration_groups (id, created_at, updated_at, deleted_at, primary_email, status, registration_date, confirmation_date, confirmation_deadline, cancellation_date, cancellation_reason, ramble_id) FROM stdin;
1	2025-09-11 15:11:47.344138+00	2025-09-11 15:11:47.344138+00	\N	vlnb12311@gmail.com	pending	2025-09-11 15:11:47.344043+00	\N	\N	\N	\N	1
2	2025-09-25 16:46:09.741069+00	2025-09-25 16:46:09.741069+00	\N	ecologique.pro@gmail.com	pending	2025-09-25 16:46:09.740607+00	\N	\N	\N	\N	13
3	2025-09-28 18:15:07.365187+00	2025-09-28 18:15:07.365187+00	\N	Vandepeutte.coline@hotmail.fr	pending	2025-09-28 18:15:07.364821+00	\N	\N	\N	\N	14
5	2025-10-01 05:25:04.071057+00	2025-10-01 05:25:04.071057+00	\N	yannick.robert35@gmail.com	pending	2025-10-01 05:25:04.070712+00	\N	\N	\N	\N	8
6	2025-10-01 09:46:43.604346+00	2025-10-01 09:46:43.604346+00	\N	clotilde.philippe@univ-rennes.fr	pending	2025-10-01 09:46:43.60397+00	\N	\N	\N	\N	2
7	2025-10-01 09:47:42.876477+00	2025-10-01 09:47:42.876477+00	\N	francoise.razanamaro@univ-rennes.fr	pending	2025-10-01 09:47:42.876187+00	\N	\N	\N	\N	6
8	2025-10-01 16:25:00.230305+00	2025-10-01 16:25:00.230305+00	\N	shauezarx@mozmail.com	pending	2025-10-01 16:25:00.229945+00	\N	\N	\N	\N	15
10	2025-10-01 19:28:56.246261+00	2025-10-01 19:28:56.246261+00	\N	jeanyvesnath@gmail.com	pending	2025-10-01 19:28:56.245888+00	\N	\N	\N	\N	17
4	2025-09-30 11:18:17.1468+00	2025-09-30 11:18:17.1468+00	\N	julientrubat86@gmail.com	pending	2025-09-30 11:18:17.146458+00	\N	\N	\N	\N	1
11	2025-10-02 13:49:50.155197+00	2025-10-02 13:49:50.155197+00	\N	catheline.barbara@gmail.com	pending	2025-10-02 13:49:50.154855+00	\N	\N	\N	\N	14
12	2025-10-03 08:20:04.84678+00	2025-10-03 08:20:04.84678+00	\N	bezier.marion@outlook.fr	pending	2025-10-03 08:20:04.846416+00	\N	\N	\N	\N	14
13	2025-10-03 12:08:10.598268+00	2025-10-03 12:08:10.598268+00	\N	guyot.rebecca@gmail.com	pending	2025-10-03 12:08:10.597858+00	\N	\N	\N	\N	15
14	2025-10-03 19:31:11.68276+00	2025-10-03 19:31:11.68276+00	\N	pauline.trion@gmail.com	pending	2025-10-03 19:31:11.682417+00	\N	\N	\N	\N	14
15	2025-10-04 19:25:30.467625+00	2025-10-04 19:25:30.467625+00	\N	gaelletwk1@yahoo.fr	pending	2025-10-04 19:25:30.467371+00	\N	\N	\N	\N	17
16	2025-10-05 10:59:30.057392+00	2025-10-05 10:59:30.057392+00	\N	erwan.vieville@gmail.com	pending	2025-10-05 10:59:30.057033+00	\N	\N	\N	\N	15
17	2025-10-05 14:42:12.966786+00	2025-10-05 14:42:12.966786+00	\N	canellethomas@gmail.com	pending	2025-10-05 14:42:12.966704+00	\N	\N	\N	\N	2
18	2025-10-07 10:18:06.789492+00	2025-10-07 10:18:06.789492+00	\N	lennysaxo@gmail.com	pending	2025-10-07 10:18:06.789203+00	\N	\N	\N	\N	2
19	2025-10-12 09:22:56.354391+00	2025-10-12 09:22:56.354391+00	\N	romainprevosteau@gmail.com	pending	2025-10-12 09:22:56.354069+00	\N	\N	\N	\N	12
20	2025-10-12 10:05:10.130412+00	2025-10-12 10:05:10.130412+00	\N	nfeurgard@gmail.com	pending	2025-10-12 10:05:10.129919+00	\N	\N	\N	\N	7
21	2025-10-15 07:36:22.427779+00	2025-10-15 07:36:22.427779+00	\N	nathalie.strugalski@hotmail.fr	pending	2025-10-15 07:36:22.427682+00	\N	\N	\N	\N	12
22	2025-10-20 08:11:18.562114+00	2025-10-20 08:11:18.562114+00	\N	daheronf@gmail.com	confirmed	2025-10-20 08:11:18.561684+00	\N	\N	\N	\N	6
23	2025-10-20 08:32:03.609548+00	2025-10-20 08:32:03.609548+00	\N	daheronf@gmail.com	confirmed	2025-10-20 08:32:03.609226+00	\N	\N	\N	\N	7
24	2025-10-20 17:42:23.741099+00	2025-10-20 17:42:23.741099+00	\N	nadegecorbe3576@gmail.com	pending	2025-10-20 17:42:23.740583+00	\N	\N	\N	\N	10
25	2025-10-20 18:43:48.405083+00	2025-10-20 18:43:48.405083+00	\N	pyfab.balcon@wanadoo.fr	waiting_list	2025-10-20 18:43:48.404585+00	\N	\N	\N	\N	13
26	2025-10-20 20:39:57.398546+00	2025-10-20 20:39:57.398546+00	\N	ecologique.pro@gmail.com	pending	2025-10-20 20:39:57.398454+00	\N	\N	\N	\N	8
27	2025-10-24 21:18:33.372717+00	2025-10-24 21:18:33.372717+00	\N	christellehuiban@gmail.com	pending	2025-10-24 21:18:33.372317+00	\N	\N	\N	\N	12
28	2025-11-07 10:03:14.530723+00	2025-11-07 10:03:14.530723+00	\N	hillion.c@gmail.com	pending	2025-11-07 10:03:14.530342+00	\N	\N	\N	\N	25
29	2025-11-08 14:32:37.788794+00	2025-11-08 14:32:37.788794+00	\N	vincent.virginie@gmail.com	pending	2025-11-08 14:32:37.78851+00	\N	\N	\N	\N	23
30	2025-11-08 17:29:07.205668+00	2025-11-08 17:29:07.205668+00	\N	emilie.gillier@lilo.org	pending	2025-11-08 17:29:07.205305+00	\N	\N	\N	\N	21
31	2025-11-14 16:59:34.352236+00	2025-11-14 16:59:34.352236+00	\N	charlotte.fourchon@gmail.com	pending	2025-11-14 16:59:34.351864+00	\N	\N	\N	\N	21
32	2026-04-18 17:05:18.254875+00	2026-04-18 17:05:18.254875+00	\N	maeva.cadeau@orange.fr	pending	2026-04-18 17:05:18.254574+00	\N	\N	\N	\N	29
33	2026-04-24 11:41:10.498474+00	2026-04-24 11:41:10.498474+00	\N	maeva.cadeau@orange.fr	pending	2026-04-24 11:41:10.498126+00	\N	\N	\N	\N	31
34	2026-05-07 18:13:01.315233+00	2026-05-07 18:13:01.315233+00	\N	yann@linuxconsole.org	confirmed	2026-05-07 18:13:01.314853+00	\N	\N	\N	\N	31
35	2026-06-15 13:37:53.326598+00	2026-06-15 13:37:53.326598+00	\N	maeva.cadeau@orange.fr 	pending	2026-06-15 13:37:53.325849+00	\N	\N	\N	\N	37
36	2026-06-17 08:37:50.376682+00	2026-06-17 08:37:50.376682+00	\N	vielderennes@orange.fr	pending	2026-06-17 08:37:50.376216+00	\N	\N	\N	\N	35
37	2026-06-18 08:00:32.366975+00	2026-06-18 08:00:32.366975+00	\N	maeva.cadeau@orange.fr	confirmed	2026-06-18 08:00:32.366476+00	\N	\N	\N	\N	37
\.


--
-- Data for Name: ramble_registrations; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.ramble_registrations (id, created_at, updated_at, deleted_at, first_name, last_name, email, phone, status, registration_date, confirmation_date, confirmation_deadline, cancellation_date, cancellation_reason, ramble_id, user_id, group_id) FROM stdin;
30	2025-10-01 10:31:10.592037+00	2025-10-22 08:23:31.816136+00	\N	xavier	even-cuilerier	laulo561@gmail.com	\N	cancelled	2025-10-01 10:31:10.591703+00	\N	2025-10-23 08:00:00+00	2025-10-22 08:23:31.815699+00	No confirmation before deadline	8	22	\N
36	2025-10-01 19:04:54.610069+00	2025-10-06 13:16:28.396203+00	\N	Charles	KERSUAL	kersual35@gmail.com	06 23 20 70 54	cancelled	2025-10-01 19:04:54.609664+00	\N	2025-10-07 12:30:00+00	2025-10-06 13:16:28.394697+00	No confirmation before deadline	2	27	\N
29	2025-10-01 10:28:52.254941+00	2025-10-20 08:06:00.758952+00	\N	Philippe	cuillerier	stephylou35@hotmail.fr	0675304855	cancelled	2025-10-01 10:28:52.254609+00	\N	2025-10-21 08:00:00+00	2025-10-20 08:06:00.758675+00	No confirmation before deadline	6	21	\N
55	2025-10-03 15:53:36.490479+00	2025-10-17 08:02:53.954378+00	\N	fabienne	balcon	pyfab.balcon@wanadoo.fr	0609948591	confirmed	2025-10-03 15:53:36.490418+00	2025-10-17 08:02:53.953981+00	2025-10-19 12:30:00+00	\N	\N	5	41	\N
53	2025-10-03 12:08:10.604632+00	2025-10-08 05:52:46.722748+00	\N	Julie	Glemot	julieglemot5@gmail.com	\N	confirmed	2025-10-03 12:08:10.604472+00	2025-10-08 05:52:46.720758+00	2025-10-10 17:30:00+00	\N	\N	15	39	13
9	2025-09-28 18:15:07.370402+00	2025-10-08 17:35:21.529026+00	\N	coline	vandepeutte	Vandepeutte.coline@hotmail.fr	0683619333	cancelled	2025-09-28 18:15:07.370237+00	\N	2025-10-09 17:30:00+00	2025-10-08 17:35:21.526842+00	No confirmation before deadline	14	5	3
10	2025-09-28 18:15:07.373245+00	2025-10-08 17:35:21.534058+00	\N	hugo	Chapel 	vandepeutte.coline@hotmail.fr	\N	cancelled	2025-09-28 18:15:07.373103+00	\N	2025-10-09 17:30:00+00	2025-10-08 17:35:21.532785+00	No confirmation before deadline	14	6	3
1	2025-09-11 15:11:47.346937+00	2025-10-02 13:18:12.91171+00	\N	vera	lor	vlnb12311@gmail.com	\N	cancelled	2025-09-11 15:11:47.346857+00	\N	2025-10-03 12:30:00+00	2025-10-02 13:18:12.910324+00	No confirmation before deadline	1	2	1
2	2025-09-11 15:11:47.348444+00	2025-10-02 13:18:12.915081+00	\N	vic	den	vlnb12311@gmail.com	\N	cancelled	2025-09-11 15:11:47.348371+00	\N	2025-10-03 12:30:00+00	2025-10-02 13:18:12.914325+00	No confirmation before deadline	1	2	1
26	2025-10-01 09:46:43.610913+00	2025-10-06 13:16:28.401025+00	\N	Sacha	Pérocheau Arnaud 	sacha.pa@free.fr	\N	cancelled	2025-10-01 09:46:43.61078+00	\N	2025-10-07 12:30:00+00	2025-10-06 13:16:28.3996+00	No confirmation before deadline	2	18	6
48	2025-10-03 08:20:04.852596+00	2025-10-08 17:35:21.536753+00	\N	Marion 	Bezier 	bezier.marion@outlook.fr	0634983715	cancelled	2025-10-03 08:20:04.852408+00	\N	2025-10-09 17:30:00+00	2025-10-08 17:35:21.535988+00	No confirmation before deadline	14	34	12
50	2025-10-03 11:13:10.858864+00	2025-10-03 11:13:10.858864+00	\N	Anne-Marie	Guillot	amarie.guillot@laposte.net	\N	pending	2025-10-03 11:13:10.858795+00	\N	\N	\N	\N	1	36	\N
25	2025-10-01 09:46:43.607978+00	2025-10-06 07:18:35.151575+00	\N	Clotilde 	Philippe 	clotilde.philippe@univ-rennes.fr	0667339738	confirmed	2025-10-01 09:46:43.607744+00	2025-10-06 07:18:35.14911+00	2025-10-07 12:30:00+00	\N	\N	2	17	6
11	2025-09-29 06:17:58.56252+00	2025-10-06 10:16:59.497323+00	\N	Soizic	POILVE	soizzzm@yahoo.fr	0634513252	confirmed	2025-09-29 06:17:58.562363+00	2025-10-06 10:16:59.494353+00	2025-10-07 12:30:00+00	\N	\N	2	7	\N
15	2025-09-30 11:18:17.153615+00	2025-10-02 08:18:12.8239+00	\N	Julien 	TRUBAT 	julientrubat86@gmail.com	0625927852	cancelled	2025-09-30 11:18:17.15338+00	\N	2025-10-03 08:00:00+00	2025-10-02 08:18:12.822412+00	No confirmation before deadline	1	10	4
16	2025-09-30 11:18:17.155864+00	2025-10-02 08:18:12.82949+00	\N	Perrine 	Oltra 	oltraperrine@gmail.com	\N	cancelled	2025-09-30 11:18:17.155815+00	\N	2025-10-03 08:00:00+00	2025-10-02 08:18:12.828125+00	No confirmation before deadline	1	11	4
17	2025-09-30 21:45:49.629632+00	2025-10-02 13:18:12.917263+00	\N	stefania	porcelli	s.porcelli92@gmail.com	0769558635	cancelled	2025-09-30 21:45:49.629476+00	\N	2025-10-03 12:30:00+00	2025-10-02 13:18:12.916643+00	No confirmation before deadline	1	13	\N
24	2025-10-01 06:56:19.036316+00	2025-10-02 13:18:12.919097+00	\N	Jean-Pierre	Guillot	jpg.guillot@gmail.com	0633740654	cancelled	2025-10-01 06:56:19.035982+00	\N	2025-10-03 12:30:00+00	2025-10-02 13:18:12.91848+00	No confirmation before deadline	1	16	\N
49	2025-10-03 08:20:04.85421+00	2025-10-08 17:35:21.538927+00	\N	Quentin 	Bellamy 	walker-35@hotmail.fr	\N	cancelled	2025-10-03 08:20:04.854166+00	\N	2025-10-09 17:30:00+00	2025-10-08 17:35:21.538129+00	No confirmation before deadline	14	35	12
51	2025-10-03 12:01:54.388317+00	2025-10-08 08:02:21.9494+00	\N	Edith	BLIN	edith.blin@inria.fr	06 84 17 81 81	confirmed	2025-10-03 12:01:54.388099+00	2025-10-08 08:02:21.948368+00	2025-10-10 17:30:00+00	\N	\N	15	37	\N
18	2025-10-01 04:21:56.780258+00	2025-10-09 17:35:21.523466+00	\N	anne	marchand	marchand.anne@wanadoo.fr	0681265968	cancelled	2025-10-01 04:21:56.779869+00	\N	2025-10-10 17:30:00+00	2025-10-09 17:35:21.523169+00	No confirmation before deadline	15	14	\N
31	2025-10-01 16:25:00.234162+00	2025-10-09 17:35:21.524676+00	\N	Julie	ANBERREE	shauezarx@mozmail.com	\N	cancelled	2025-10-01 16:25:00.233809+00	\N	2025-10-10 17:30:00+00	2025-10-09 17:35:21.524444+00	No confirmation before deadline	15	23	8
32	2025-10-01 16:25:00.235908+00	2025-10-09 17:35:21.525596+00	\N	Thomas	ANBERREE	thomas.anberree@gmail.com	\N	cancelled	2025-10-01 16:25:00.23586+00	\N	2025-10-10 17:30:00+00	2025-10-09 17:35:21.525341+00	No confirmation before deadline	15	24	8
42	2025-10-02 12:49:19.731811+00	2025-10-20 13:23:31.821131+00	\N	Isabelle	Lins	isabelle.lins@univ-rennes.fr	0771122770	cancelled	2025-10-02 12:49:19.731627+00	\N	2025-10-21 12:30:00+00	2025-10-20 13:23:31.819806+00	No confirmation before deadline	7	31	\N
38	2025-10-01 19:28:56.250101+00	2025-10-15 07:23:23.251388+00	\N	jean yves	dabo	jeanyvesnath@gmail.com	0643802854	confirmed	2025-10-01 19:28:56.249899+00	2025-10-15 07:23:23.249243+00	2025-10-17 07:00:00+00	\N	\N	17	28	10
28	2025-10-01 09:47:42.882686+00	2025-10-20 08:06:00.75767+00	\N	Damien	BARBEDETTE	barbedettedamien@yahoo.fr	\N	cancelled	2025-10-01 09:47:42.882526+00	\N	2025-10-21 08:00:00+00	2025-10-20 08:06:00.757231+00	No confirmation before deadline	6	20	7
40	2025-10-02 08:19:41.273318+00	2025-10-19 07:23:34.319897+00	\N	Mathieu	BELLEC	ann.dauphin@laposte.net	0750038615	cancelled	2025-10-02 08:19:41.273121+00	\N	2025-10-21 08:00:00+00	2025-10-19 07:23:34.318054+00	\N	6	30	\N
20	2025-10-01 05:25:04.076443+00	2025-10-21 07:43:27.178482+00	\N	Yannick 	Robert 	yannick.robert35@gmail.com	0679821689	confirmed	2025-10-01 05:25:04.076084+00	2025-10-21 07:43:27.176285+00	2025-10-23 08:00:00+00	\N	\N	8	15	5
27	2025-10-01 09:47:42.880009+00	2025-10-13 12:37:22.871531+00	\N	Françoise	RAZANAMARO	francoise.razanamaro@univ-rennes.fr	0688103431	cancelled	2025-10-01 09:47:42.879814+00	\N	\N	2025-10-13 12:37:22.8685+00	Cela tombe pendant mes vacances où je suis en déplacement.	6	19	7
21	2025-10-01 05:25:04.078888+00	2025-10-21 07:43:24.64895+00	\N	aline	robert	yannick.robert35@gmail.com	\N	confirmed	2025-10-01 05:25:04.078691+00	2025-10-21 07:43:24.64836+00	2025-10-23 08:00:00+00	\N	\N	8	15	5
22	2025-10-01 05:25:04.080283+00	2025-10-21 07:43:18.936475+00	\N	lucas	robert	yannick.robert35@gmail.com	\N	confirmed	2025-10-01 05:25:04.080085+00	2025-10-21 07:43:18.934265+00	2025-10-23 08:00:00+00	\N	\N	8	15	5
54	2025-10-03 15:13:00.80289+00	2025-10-22 08:23:32.457247+00	\N	Béatrice 	Rabault	beatrice.rabault@gmail.com	0676057291	cancelled	2025-10-03 15:13:00.802673+00	\N	2025-10-23 08:00:00+00	2025-10-22 08:23:32.456728+00	No confirmation before deadline	8	40	\N
35	2025-10-01 19:04:52.920222+00	2025-10-27 09:23:31.816329+00	\N	Jeanne	LICHOU	jeanne.lichou@gmail.com	0623682355	cancelled	2025-10-01 19:04:52.919818+00	\N	2025-10-28 09:00:00+00	2025-10-27 09:23:31.816036+00	No confirmation before deadline	10	25	\N
23	2025-10-01 05:25:04.081801+00	2025-10-21 07:43:16.473719+00	\N	chkoe	robert	yannick.robert35@gmail.com	\N	confirmed	2025-10-01 05:25:04.081589+00	2025-10-21 07:43:16.471488+00	2025-10-23 08:00:00+00	\N	\N	8	15	5
13	2025-09-29 14:46:53.066021+00	2025-10-24 05:04:16.728494+00	\N	Lolita	COZETTE 	lolia.cozette@gmail.com	0640395643	confirmed	2025-09-29 14:46:53.06583+00	2025-10-24 05:04:16.726056+00	2025-10-26 13:30:00+00	\N	\N	9	8	\N
47	2025-10-02 18:27:29.736658+00	2025-10-24 16:48:43.936765+00	\N	Elodie	Octau	elodie.octau@gmail.com	0672340773	confirmed	2025-10-02 18:27:29.736439+00	2025-10-24 16:48:43.934732+00	2025-10-26 13:30:00+00	\N	\N	9	33	\N
39	2025-10-01 19:28:56.252091+00	2025-10-15 12:27:18.0777+00	\N	nathalie	muller	muller.nath8@gmail.com	0687340223	confirmed	2025-10-01 19:28:56.252036+00	2025-10-15 12:27:18.07557+00	2025-10-17 07:00:00+00	\N	\N	17	29	10
19	2025-10-01 04:24:20.021806+00	2025-10-16 07:14:37.939682+00	\N	anne	marchand 	marchand.anne@wanadoo.fr	\N	cancelled	2025-10-01 04:24:20.021456+00	\N	2025-10-17 07:00:00+00	2025-10-16 07:14:37.935864+00	Un changement de planning me contraint à annuler ma participation à  la sortie de samedi 18/10 à Careil.\nJe vous souhaite une belle balade	17	14	\N
37	2025-10-01 19:05:41.656179+00	2025-10-27 09:23:31.817921+00	\N	Delphine	BECHETOILLE	delphinebechetoille@gmail.com	\N	cancelled	2025-10-01 19:05:41.655777+00	\N	2025-10-28 09:00:00+00	2025-10-27 09:23:31.817199+00	No confirmation before deadline	10	26	\N
12	2025-09-29 08:12:56.10911+00	2025-10-28 07:08:30.869576+00	\N	coline	Vandepeutte 	Vandepeutte.coline@hotmail.fr	0683619333	confirmed	2025-09-29 08:12:56.108718+00	2025-10-28 07:08:30.86753+00	2025-10-30 09:00:00+00	\N	\N	12	5	\N
56	2025-10-03 16:46:20.556921+00	2025-10-03 16:46:20.556921+00	\N	stephanie	le labousse	stephanie.lelabousse@gmail.com	0681781953	pending	2025-10-03 16:46:20.556688+00	\N	\N	\N	\N	1	42	\N
57	2025-10-03 16:47:24.206437+00	2025-10-03 16:47:24.206437+00	\N	christophe	berthault	ch.berthault@gmail.com	\N	pending	2025-10-03 16:47:24.206198+00	\N	\N	\N	\N	1	43	\N
52	2025-10-03 12:08:10.601805+00	2025-10-08 08:02:35.713583+00	\N	Rébecca	GUYOT	guyot.rebecca@gmail.com	\N	confirmed	2025-10-03 12:08:10.601595+00	2025-10-08 08:02:35.711201+00	2025-10-10 17:30:00+00	\N	\N	15	38	13
76	2025-10-07 10:18:06.796568+00	2025-10-13 13:29:34.964794+00	\N	Lenny 	Legrand 	lennysaxo@gmail.com	0671187096	cancelled	2025-10-07 10:18:06.796159+00	\N	\N	2025-10-13 13:29:34.962363+00	malade	2	58	18
77	2025-10-07 10:18:06.799665+00	2025-10-13 13:29:42.540387+00	\N	Chabrol	Christian	lennysaxo@gmail.com	\N	cancelled	2025-10-07 10:18:06.7995+00	\N	\N	2025-10-13 13:29:42.539806+00	Malade	2	58	18
69	2025-10-05 14:42:12.970014+00	2025-10-07 07:23:31.468292+00	\N	Charlotte	Rondel	rondel.charlotte@gmail.com	\N	confirmed	2025-10-05 14:42:12.96996+00	2025-10-07 07:23:31.467829+00	2025-10-07 12:30:00+00	\N	\N	2	53	17
68	2025-10-05 14:42:12.968732+00	2025-10-07 07:23:38.681136+00	\N	Canelle	Thomas	canellethomas@gmail.com	0667087378	confirmed	2025-10-05 14:42:12.968677+00	2025-10-07 07:23:38.679249+00	2025-10-07 12:30:00+00	\N	\N	2	52	17
64	2025-10-05 10:59:30.061441+00	2025-10-09 17:35:21.526321+00	\N	Erwan 	Viéville 	erwan.vieville@gmail.com	\N	cancelled	2025-10-05 10:59:30.06125+00	\N	2025-10-10 17:30:00+00	2025-10-09 17:35:21.526045+00	No confirmation before deadline	15	48	16
65	2025-10-05 10:59:30.063388+00	2025-10-09 17:35:21.52705+00	\N	Rosalba	Viéville 	rosalba.vieville@gmail.com	\N	cancelled	2025-10-05 10:59:30.06332+00	\N	2025-10-10 17:30:00+00	2025-10-09 17:35:21.526851+00	No confirmation before deadline	15	49	16
6	2025-09-25 16:46:09.756098+00	2025-11-10 10:48:55.314692+00	\N	et4	et44	ecologique.pro@gmail.com	\N	confirmed	2025-09-25 16:46:09.755939+00	2025-11-10 10:48:55.313513+00	2025-11-10 09:30:00+00	\N	\N	13	3	2
66	2025-10-05 10:59:30.06485+00	2025-10-09 17:35:21.527697+00	\N	Yvon	Viéville 	yvon.vieville@gmail.com	\N	cancelled	2025-10-05 10:59:30.064799+00	\N	2025-10-10 17:30:00+00	2025-10-09 17:35:21.527506+00	No confirmation before deadline	15	50	16
43	2025-10-02 13:49:50.158646+00	2025-10-07 08:12:58.12404+00	\N	jocelyne	giraux	catheline.barbara@gmail.com	0685423249	confirmed	2025-10-02 13:49:50.158472+00	2025-10-07 08:12:58.123044+00	2025-10-09 17:30:00+00	\N	\N	14	32	11
45	2025-10-02 13:49:50.161069+00	2025-10-07 08:12:53.593848+00	\N	tristan	Catheline-giraux	catheline.barbara@gmail.com	0685423249	confirmed	2025-10-02 13:49:50.160997+00	2025-10-07 08:12:53.593489+00	2025-10-09 17:30:00+00	\N	\N	14	32	11
46	2025-10-02 13:49:50.161558+00	2025-10-07 08:12:50.512743+00	\N	alicia	Da costa	catheline.barbara@gmail.com	0685423249	confirmed	2025-10-02 13:49:50.161507+00	2025-10-07 08:12:50.512356+00	2025-10-09 17:30:00+00	\N	\N	14	32	11
44	2025-10-02 13:49:50.160137+00	2025-10-07 08:12:56.06173+00	\N	jean Louis 	giraux	catheline.barbara@gmail.com	0685423249	confirmed	2025-10-02 13:49:50.159943+00	2025-10-07 08:12:56.061001+00	2025-10-09 17:30:00+00	\N	\N	14	32	11
60	2025-10-03 19:31:11.690302+00	2025-10-07 08:36:12.615348+00	\N	Liza	Simon	pauline.trion@gmail.com	\N	confirmed	2025-10-03 19:31:11.690237+00	2025-10-07 08:36:12.614296+00	2025-10-09 17:30:00+00	\N	\N	14	44	14
59	2025-10-03 19:31:11.689666+00	2025-10-07 08:36:14.476849+00	\N	Pierre	Briand	pauline.trion@gmail.com	\N	confirmed	2025-10-03 19:31:11.689574+00	2025-10-07 08:36:14.476378+00	2025-10-09 17:30:00+00	\N	\N	14	44	14
58	2025-10-03 19:31:11.688521+00	2025-10-07 08:36:16.090687+00	\N	Pauline	Trion	pauline.trion@gmail.com	0633867441	confirmed	2025-10-03 19:31:11.688332+00	2025-10-07 08:36:16.088932+00	2025-10-09 17:30:00+00	\N	\N	14	44	14
81	2025-10-11 17:40:12.397767+00	2025-10-11 17:40:12.397767+00	\N	Victor	DENIS	victordenis01@gmail.com	\N	confirmed	2025-10-11 17:40:12.397348+00	2025-10-11 17:40:12.397348+00	\N	\N	\N	20	62	\N
72	2025-10-07 06:38:16.241721+00	2025-10-07 06:38:16.241721+00	\N	Maéva	CADEAU	maeva.cadeau@orange.fr	0619980199	pending	2025-10-07 06:38:16.24162+00	\N	\N	\N	\N	18	45	\N
67	2025-10-05 14:27:26.396318+00	2025-10-17 05:07:00.266387+00	\N	Renaud	Delannay	r.delannay@orange.fr	0608151402	confirmed	2025-10-05 14:27:26.395971+00	2025-10-17 05:07:00.264154+00	2025-10-19 12:30:00+00	\N	\N	5	51	\N
86	2025-10-13 17:22:50.714327+00	2025-10-19 17:44:54.565779+00	\N	Laura 	Giommi 	laura.giommi@live.fr	0678461720	confirmed	2025-10-13 17:22:50.714009+00	2025-10-19 17:44:54.563295+00	2025-10-21 08:00:00+00	\N	\N	6	66	\N
74	2025-10-07 08:12:46.870842+00	2025-10-20 13:23:31.826486+00	\N	martine	piel	mart.piel@orange.fr	0681642064	cancelled	2025-10-07 08:12:46.87045+00	\N	2025-10-21 12:30:00+00	2025-10-20 13:23:31.825036+00	No confirmation before deadline	7	56	\N
87	2025-10-13 17:24:48.723927+00	2025-10-19 17:44:52.321835+00	\N	Solène 	Barbé 	solene.barbe@hotmail.fr	\N	confirmed	2025-10-13 17:24:48.723744+00	2025-10-19 17:44:52.318934+00	2025-10-21 08:00:00+00	\N	\N	6	67	\N
85	2025-10-12 10:05:10.14125+00	2025-10-22 06:43:14.672332+00	\N	Lara	Schembri	isabelle.brendlen@icloud.com	0682104561	cancelled	2025-10-12 10:05:10.141073+00	2025-10-19 15:32:26.083803+00	2025-10-21 12:30:00+00	2025-10-22 06:43:14.671877+00	impossibilité d'aller jusqu'à liffre ; on pensait que c'était à paimpont. \nExcusez-nous.	7	65	20
84	2025-10-12 10:05:10.136998+00	2025-10-22 07:10:11.01565+00	\N	Nina	Feurgard	nfeurgard@gmail.com	\N	cancelled	2025-10-12 10:05:10.136652+00	2025-10-20 11:47:10.610599+00	2025-10-21 12:30:00+00	2025-10-22 07:10:11.015043+00	\N	7	64	20
78	2025-10-07 14:06:22.263836+00	2025-10-21 15:32:42.064951+00	\N	iven	Le Louedec	iven.lelouedec@gmail.com	0649810682	confirmed	2025-10-07 14:06:22.263452+00	2025-10-21 15:32:42.064363+00	2025-10-23 08:00:00+00	\N	\N	8	59	\N
95	2025-10-15 16:49:09.782022+00	2025-10-24 16:45:09.840039+00	\N	David 	Octau 	davoctau@gmail.com	0629455581	confirmed	2025-10-15 16:49:09.781631+00	2025-10-24 16:45:09.836868+00	2025-10-26 13:30:00+00	\N	\N	9	73	\N
70	2025-10-06 16:04:32.487607+00	2025-10-07 16:14:08.342296+00	\N	adeline	Léon 	adline.leon@gmail.com	\N	cancelled	2025-10-06 16:04:32.487557+00	\N	\N	2025-10-07 16:14:08.341879+00	\N	10	54	\N
62	2025-10-04 19:25:30.471278+00	2025-10-15 06:43:48.865714+00	\N	gaelle	tworkowski	gaelletwk1@yahoo.fr	0673009969	confirmed	2025-10-04 19:25:30.471048+00	2025-10-15 06:43:48.863638+00	2025-10-17 07:00:00+00	\N	\N	17	46	15
63	2025-10-04 19:25:30.474559+00	2025-10-15 06:36:28.429801+00	\N	mickael	dardaillon	mickael.dardaillon@gmail.com	\N	confirmed	2025-10-04 19:25:30.474396+00	2025-10-15 06:36:28.427295+00	2025-10-17 07:00:00+00	\N	\N	17	47	15
90	2025-10-13 22:02:19.169033+00	2025-10-15 11:50:29.89873+00	\N	Apolline	Privat	apollineprivat@gmail.com	\N	confirmed	2025-10-13 22:02:19.168832+00	2025-10-15 11:50:29.895208+00	2025-10-17 07:00:00+00	\N	\N	17	70	\N
82	2025-10-12 09:22:56.363482+00	2025-10-28 07:42:21.173654+00	\N	Romain	Prevosteau 	romainprevosteau@gmail.com	0628451289	confirmed	2025-10-12 09:22:56.363288+00	2025-10-28 07:42:21.171831+00	2025-10-30 09:00:00+00	\N	\N	12	63	19
91	2025-10-14 19:48:53.320252+00	2025-11-23 16:18:16.369393+00	\N	Alexia	Muzas	axa_mzs_work@outlook.fr	0625358869	confirmed	2025-10-14 19:48:53.319751+00	2025-11-23 16:18:16.36748+00	2025-10-30 09:00:00+00	\N	\N	12	71	\N
7	2025-09-25 16:46:09.757986+00	2025-11-10 10:48:55.310444+00	\N	et5	et55	ecologique.pro@gmail.com	\N	confirmed	2025-09-25 16:46:09.757815+00	2025-11-10 10:48:55.308491+00	2025-11-10 09:30:00+00	\N	\N	13	3	2
3	2025-09-25 16:46:09.750078+00	2025-11-10 10:48:55.321692+00	\N	Vera	Lorenzetti	ecologique.pro@gmail.com	\N	confirmed	2025-09-25 16:46:09.749863+00	2025-11-10 10:48:55.321269+00	2025-11-10 09:30:00+00	\N	\N	13	3	2
89	2025-10-13 22:00:05.96678+00	2025-10-24 10:00:26.798478+00	\N	Gaël Yann 	Rubin 	gael-yann@hotmail.fr	0626877254	confirmed	2025-10-13 22:00:05.966577+00	2025-10-24 10:00:26.795971+00	2025-10-26 13:30:00+00	\N	\N	9	69	\N
71	2025-10-06 16:06:03.585686+00	2025-10-28 05:45:19.526516+00	\N	adeline	Léon 	adline.leon@gmail.com	\N	confirmed	2025-10-06 16:06:03.585244+00	2025-10-28 05:45:19.524295+00	2025-10-30 09:00:00+00	\N	\N	12	54	\N
83	2025-10-12 09:22:56.36616+00	2025-10-28 07:42:17.412255+00	\N	Fanny	Prevosteau 	romainprevosteau@gmail.com	\N	confirmed	2025-10-12 09:22:56.365976+00	2025-10-28 07:42:17.410202+00	2025-10-30 09:00:00+00	\N	\N	12	63	19
92	2025-10-15 07:36:22.429452+00	2025-10-29 07:01:11.748938+00	\N	Nathalie 	Delvincourt 	nathalie.strugalski@hotmail.fr	0699526971	confirmed	2025-10-15 07:36:22.4294+00	2025-10-29 07:01:11.747007+00	2025-10-30 09:00:00+00	\N	\N	12	72	21
79	2025-10-09 18:33:00.148415+00	2025-10-16 07:20:21.94427+00	\N	cyril	pinchon	cyrilpinchon@gmail.com	0634457421	cancelled	2025-10-09 18:33:00.148313+00	\N	\N	2025-10-16 07:20:21.941708+00	Changement de programme, je viens l'après-midi	6	60	\N
61	2025-10-03 19:50:51.510219+00	2025-10-15 19:15:22.421876+00	\N	Maéva	CADEAU	maeva.cadeau@orange.fr	0619980199	confirmed	2025-10-03 19:50:51.510032+00	2025-10-15 19:15:22.42152+00	2025-10-17 07:00:00+00	\N	\N	17	45	\N
94	2025-10-15 12:12:13.147324+00	2025-10-19 06:15:25.87848+00	\N	cyril	pinchon	cyrilpinchon@gmail.com	06 34 45 74 21	confirmed	2025-10-15 12:12:13.146922+00	2025-10-19 06:15:25.878036+00	2025-10-21 12:30:00+00	\N	\N	7	60	\N
75	2025-10-07 08:15:51.329617+00	2025-10-20 13:23:31.830372+00	\N	pascal	Navarre	pascal.navarre@wanadoo.fr	0680069086	cancelled	2025-10-07 08:15:51.329538+00	\N	2025-10-21 12:30:00+00	2025-10-20 13:23:31.829118+00	No confirmation before deadline	7	57	\N
124	2025-10-22 09:48:58.620928+00	2025-10-22 09:48:58.620928+00	\N	Béatricen	Rabault	beatice.rabault@gmail.com	0676057291	confirmed	2025-10-22 09:48:58.620557+00	2025-10-22 09:48:58.620557+00	\N	\N	\N	8	83	\N
73	2025-10-07 06:57:29.466126+00	2025-10-17 06:12:00.470926+00	\N	Nadège 	Lécrivain	nadegecorbe3576@gmail.com	0674659708	confirmed	2025-10-07 06:57:29.465926+00	2025-10-17 06:12:00.468851+00	2025-10-19 12:30:00+00	\N	\N	5	55	\N
98	2025-10-19 17:52:18.689246+00	2025-10-19 18:20:28.609539+00	\N	manon	demeulenaere	demeulenaere.manon@yahoo.fr	\N	confirmed	2025-10-19 17:52:18.68893+00	2025-10-19 17:52:18.68893+00	2025-10-21 08:00:00+00	\N	\N	6	77	\N
104	2025-10-20 08:24:12.001508+00	2025-10-20 08:24:12.001508+00	\N	Victor	DENIS	victordenis01@gmail.com	\N	confirmed	2025-10-20 08:24:12.001353+00	2025-10-20 08:24:12.001353+00	\N	\N	\N	6	62	\N
105	2025-10-20 08:32:03.614192+00	2025-10-20 08:32:03.614192+00	\N	Francis	DAHERON	daheronf@gmail.com	\N	confirmed	2025-10-20 08:32:03.614015+00	2025-10-20 08:32:03.609226+00	\N	\N	\N	7	68	23
106	2025-10-20 08:32:03.617162+00	2025-10-20 08:32:03.617162+00	\N	Marie	DAHERON GAUGAIN	daheronf@gmail.com	\N	confirmed	2025-10-20 08:32:03.616968+00	2025-10-20 08:32:03.609226+00	\N	\N	\N	7	68	23
107	2025-10-20 08:32:03.619009+00	2025-10-20 08:32:03.619009+00	\N	Waël	DAHERON GAUGAIN	daheronf@gmail.com	\N	confirmed	2025-10-20 08:32:03.618793+00	2025-10-20 08:32:03.609226+00	\N	\N	\N	7	68	23
108	2025-10-20 08:32:03.620865+00	2025-10-20 08:32:03.620865+00	\N	Malika	TENEUR	malika.teneur@gmail.com	\N	confirmed	2025-10-20 08:32:03.620679+00	2025-10-20 08:32:03.609226+00	\N	\N	\N	7	78	23
97	2025-10-18 09:48:07.791724+00	2025-10-25 14:23:31.822375+00	\N	Anthony 	Ambroise 	anthonyambroise68@gmail.com	0683539476	cancelled	2025-10-18 09:48:07.791594+00	\N	2025-10-26 13:30:00+00	2025-10-25 14:23:31.820286+00	No confirmation before deadline	9	76	\N
120	2025-10-20 20:39:57.402369+00	2025-10-21 07:23:05.715396+00	\N	vera	sept	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.402327+00	2025-10-21 07:23:05.713275+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
119	2025-10-20 20:39:57.401979+00	2025-10-21 07:23:11.898117+00	\N	vera	six	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.401936+00	2025-10-21 07:23:11.894765+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
118	2025-10-20 20:39:57.401532+00	2025-10-21 07:23:15.975533+00	\N	vera	cinq	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.401469+00	2025-10-21 07:23:15.973286+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
117	2025-10-20 20:39:57.401097+00	2025-10-21 07:23:25.936365+00	\N	vera	quatre	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.401037+00	2025-10-21 07:23:25.935894+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
116	2025-10-20 20:39:57.400637+00	2025-10-21 07:23:30.045611+00	\N	vera	trois	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.400574+00	2025-10-21 07:23:30.043108+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
115	2025-10-20 20:39:57.400071+00	2025-10-21 07:23:34.403266+00	\N	vera	deux	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.399976+00	2025-10-21 07:23:34.400999+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
114	2025-10-20 20:39:57.39934+00	2025-10-21 07:23:41.525115+00	\N	vera	lor	ecologique.pro@gmail.com	\N	confirmed	2025-10-20 20:39:57.399276+00	2025-10-21 07:23:41.522868+00	2025-10-23 08:00:00+00	\N	\N	8	3	26
109	2025-10-20 10:03:27.820935+00	2025-10-21 07:46:25.802786+00	\N	david	lecocq	d_lecocq@hotmail.com	0612335399	confirmed	2025-10-20 10:03:27.820743+00	2025-10-21 07:46:25.802356+00	2025-10-23 08:00:00+00	\N	\N	8	79	\N
121	2025-10-21 11:32:11.440629+00	2025-10-21 11:32:11.440629+00	\N	estelle	clua	estelle.clua@gmail.com	0604018719	confirmed	2025-10-21 11:32:11.440269+00	2025-10-21 11:32:11.440269+00	\N	\N	\N	6	80	\N
110	2025-10-20 17:42:23.743895+00	2025-10-27 09:23:31.81925+00	\N	Lécrivain 	Nadège 	nadegecorbe3576@gmail.com	0674659708	cancelled	2025-10-20 17:42:23.743564+00	\N	2025-10-28 09:00:00+00	2025-10-27 09:23:31.818887+00	No confirmation before deadline	10	55	24
8	2025-09-25 17:32:54.821496+00	2025-11-11 17:46:46.882917+00	\N	Isabelle 	Cheval 	isabellecheval8@gmail.com	\N	confirmed	2025-09-25 17:32:54.821297+00	2025-11-11 17:46:44.619372+00	2025-11-10 09:30:00+00	\N	\N	13	4	\N
111	2025-10-20 17:42:23.74741+00	2025-10-27 09:23:31.820326+00	\N	Guénard 	Iléna	nadegecorbe3576@gmail.com	\N	cancelled	2025-10-20 17:42:23.747078+00	\N	2025-10-28 09:00:00+00	2025-10-27 09:23:31.820014+00	No confirmation before deadline	10	55	24
122	2025-10-21 11:57:00.255156+00	2025-10-23 10:56:18.450819+00	\N	diane	Schmidt	diane.schmitt3@wanadoo.fr	\N	cancelled	2025-10-21 11:57:00.254963+00	\N	\N	2025-10-23 10:56:18.448602+00	Pas dispo au final	8	81	\N
138	2025-10-29 08:47:45.31483+00	2025-10-29 08:47:45.31483+00	\N	je sais	pas 1	ecologique.pro@gmail.com	\N	confirmed	2025-10-29 08:47:45.314395+00	2025-10-29 08:47:45.314394+00	\N	\N	\N	10	3	\N
135	2025-10-26 07:42:41.780452+00	2025-10-26 07:42:41.780452+00	\N	Maëliss 	Monbon	maeliss.monbon@lilo.org	0783804201	confirmed	2025-10-26 07:42:41.780103+00	2025-10-26 07:42:41.780103+00	\N	\N	\N	9	89	\N
127	2025-10-24 09:04:08.389157+00	2025-10-24 10:16:10.783954+00	\N	Pauline	Guilbaud 	pauline-guilbaud@hotmail.fr	0689051654	confirmed	2025-10-24 09:04:08.388767+00	2025-10-24 10:16:10.781808+00	2025-10-26 13:30:00+00	\N	\N	9	85	\N
125	2025-10-23 11:11:12.483764+00	2025-10-24 11:42:23.542931+00	\N	MME	Templier	ecologique.pro@gmail.com	0785565143	confirmed	2025-10-23 11:11:12.483362+00	2025-10-24 11:42:23.540773+00	2025-10-26 13:30:00+00	\N	\N	9	3	\N
136	2025-10-26 08:37:08.308604+00	2025-10-26 08:37:08.308604+00	\N	Viki	Villeneuve	viki.villeneuve@hotmail.fr	0638781842	confirmed	2025-10-26 08:37:08.308155+00	2025-10-26 08:37:08.308155+00	\N	\N	\N	9	90	\N
130	2025-10-24 21:18:33.384396+00	2025-10-30 09:53:44.512568+00	\N	Joanne	MONTAVONT	christellehuiban@gmail.com	\N	confirmed	2025-10-24 21:18:33.384217+00	2025-10-30 09:53:44.511985+00	2025-10-30 09:00:00+00	\N	\N	12	86	27
128	2025-10-24 21:18:33.378618+00	2025-10-30 09:53:44.515662+00	\N	christelle	HUIBAN	christellehuiban@gmail.com	0649588505	confirmed	2025-10-24 21:18:33.378412+00	2025-10-30 09:53:44.515072+00	2025-10-30 09:00:00+00	\N	\N	12	86	27
80	2025-10-10 09:01:47.032517+00	2025-11-09 10:23:31.928614+00	\N	Soizic	MASSON	soizic.masson@orange.fr	07 83 33 53 52	cancelled	2025-10-10 09:01:47.032415+00	\N	2025-11-10 09:30:00+00	2025-11-09 10:23:31.927913+00	No confirmation before deadline	13	61	\N
5	2025-09-25 16:46:09.75433+00	2025-11-10 10:48:55.318532+00	\N	et3	et33	ecologique.pro@gmail.com	\N	confirmed	2025-09-25 16:46:09.754157+00	2025-11-10 10:48:55.31602+00	\N	\N	\N	13	3	2
123	2025-10-21 12:09:25.376346+00	2025-10-26 11:44:00.89848+00	\N	Titouan	Millon	titouan.millon3@gmail.com	0650040906	confirmed	2025-10-21 12:09:25.376264+00	2025-10-26 11:44:00.896027+00	2025-10-28 13:30:00+00	\N	\N	11	82	\N
93	2025-10-15 07:36:22.430232+00	2025-10-29 07:01:04.731147+00	\N	Cédric 	Delvincourt 	nathalie.strugalski@hotmail.fr	0678067704	confirmed	2025-10-15 07:36:22.430153+00	2025-10-29 07:01:04.729197+00	2025-10-30 09:00:00+00	\N	\N	12	72	21
139	2025-10-29 08:48:21.995942+00	2025-10-29 08:48:21.995942+00	\N	je sais	pas 2	vlnb12311@gmail.com	\N	confirmed	2025-10-29 08:48:21.995496+00	2025-10-29 08:48:21.995496+00	\N	\N	\N	10	2	\N
41	2025-10-02 12:25:34.351504+00	2025-11-11 17:46:02.699546+00	\N	Marie-Pierre	REMEUR	mariepierreremeur@gmail.com	\N	cancelled	2025-10-02 12:25:34.351021+00	2025-11-10 10:49:24.253711+00	2025-11-10 09:30:00+00	2025-11-11 17:46:02.697547+00	Son petit fils débarque chez elle au dernier moment et elle annule le matin même.	13	9	\N
137	2025-10-26 12:04:46.192182+00	2025-10-26 12:23:31.859844+00	\N	Victor	Blanchard	viblanchar@gmail.com	\N	confirmed	2025-10-26 12:04:46.191774+00	2025-10-26 12:04:46.191774+00	2025-10-28 09:00:00+00	\N	\N	10	91	\N
4	2025-09-25 16:46:09.75272+00	2025-11-10 10:48:55.320152+00	\N	et2	et22	ecologique.pro@gmail.com	\N	confirmed	2025-09-25 16:46:09.752516+00	2025-11-10 10:48:55.319585+00	\N	\N	\N	13	3	2
134	2025-10-25 12:46:55.126584+00	2025-11-08 06:29:35.344679+00	\N	Gaela	Cochennec	gaela_c@protonmail.com	0610491382	confirmed	2025-10-25 12:46:55.126246+00	2025-11-08 06:29:35.344278+00	2025-11-10 09:30:00+00	\N	\N	13	88	\N
96	2025-10-17 16:43:42.970758+00	2025-11-23 16:18:16.364285+00	\N	Amandine	Jegat	alexia.muzas@gmail.com	0625358869	confirmed	2025-10-17 16:43:42.970702+00	2025-11-23 16:17:53.363263+00	2025-10-30 09:00:00+00	\N	\N	12	75	\N
140	2025-10-29 12:55:23.571336+00	2025-10-29 12:55:23.571336+00	\N	Cyrille	Prevosteau	cyrille.prevosteau@sfr.fr	\N	confirmed	2025-10-29 12:55:23.571077+00	2025-10-29 12:55:23.571077+00	\N	\N	\N	12	92	\N
132	2025-10-24 21:18:33.387084+00	2025-10-30 09:53:44.508303+00	\N	Bérénice	MONTAVONT	christellehuiban@gmail.com	\N	confirmed	2025-10-24 21:18:33.387014+00	2025-10-30 09:53:44.507056+00	2025-10-30 09:00:00+00	\N	\N	12	86	27
131	2025-10-24 21:18:33.385917+00	2025-10-30 09:53:44.51091+00	\N	Charlotte	MONTAVONT	christellehuiban@gmail.com	\N	confirmed	2025-10-24 21:18:33.385708+00	2025-10-30 09:53:44.510149+00	2025-10-30 09:00:00+00	\N	\N	12	86	27
129	2025-10-24 21:18:33.382729+00	2025-10-30 09:53:44.514+00	\N	Nicolas	MONTAVONT	nicolas@montavont.net	\N	confirmed	2025-10-24 21:18:33.382565+00	2025-10-30 09:53:44.5135+00	2025-10-30 09:00:00+00	\N	\N	12	87	27
149	2025-11-08 17:29:07.213381+00	2025-11-20 06:37:22.201156+00	\N	Victor 	Gaudard 	victor.15156@gmail.com	\N	confirmed	2025-11-08 17:29:07.213196+00	2025-11-20 06:37:22.198654+00	2025-11-22 09:00:00+00	\N	\N	21	100	30
144	2025-11-07 20:15:02.17345+00	2025-11-20 11:25:02.386053+00	\N	marie	Bonnardot 	Marie.bonnardot@outlook.com	\N	confirmed	2025-11-07 20:15:02.173375+00	2025-11-20 11:25:02.382756+00	2025-11-22 09:00:00+00	\N	\N	21	96	\N
148	2025-11-08 17:29:07.209755+00	2025-11-20 12:08:08.252186+00	\N	Émilie 	Gillier 	emilie.gillier@lilo.org	\N	confirmed	2025-11-08 17:29:07.209594+00	2025-11-20 12:08:08.251545+00	2025-11-22 09:00:00+00	\N	\N	21	99	30
133	2025-10-25 09:55:54.014885+00	2025-10-30 20:11:54.387596+00	\N	Sean 	Compton	beatrice.rabault@gmail.com	0676057291	cancelled	2025-10-25 09:55:54.014482+00	\N	\N	2025-10-30 20:11:54.384621+00	\N	13	40	\N
153	2025-11-14 16:59:34.356063+00	2025-11-20 12:56:25.498053+00	\N	charlotte 	fourchon	charlotte.fourchon@gmail.com	\N	confirmed	2025-11-14 16:59:34.355888+00	2025-11-20 12:56:25.49598+00	2025-11-22 09:00:00+00	\N	\N	21	103	31
151	2025-11-09 20:36:50.945747+00	2025-11-21 09:23:31.916042+00	\N	Marie-Florine 	Dambakizi 	mf.dambakizi@gmail.com	0767935251	cancelled	2025-11-09 20:36:50.945695+00	\N	2025-11-22 09:00:00+00	2025-11-21 09:23:31.914538+00	No confirmation before deadline	21	101	\N
126	2025-10-23 11:36:24.957632+00	2025-11-07 14:29:06.754005+00	\N	mathilde	LEFRERE	cauchy.lucie@hotmail.fr	\N	cancelled	2025-10-23 11:36:24.95728+00	\N	\N	2025-11-07 14:29:06.751917+00	Cadeau pour personne non disponible	13	84	\N
112	2025-10-20 18:43:48.407481+00	2025-11-08 00:23:37.758574+00	\N	fabienne	balcon	pyfab.balcon@wanadoo.fr	0609948591	confirmed	2025-10-20 18:43:48.407296+00	2025-11-07 14:26:28.370911+00	2025-11-10 09:30:00+00	\N	\N	13	41	25
113	2025-10-20 18:43:48.408988+00	2025-11-08 00:23:38.871894+00	\N	pierre yves	balcon	pyfab.balcon@wanadoo.fr	0670750569	confirmed	2025-10-20 18:43:48.408805+00	2025-11-07 14:26:33.755718+00	2025-11-10 09:30:00+00	\N	\N	13	41	25
142	2025-11-07 10:03:14.538117+00	2025-11-12 08:43:55.483691+00	\N	Clémence	Hillion	hillion.c@gmail.com	0680361083	confirmed	2025-11-07 10:03:14.537951+00	2025-11-12 08:43:55.481674+00	2025-11-14 08:00:00+00	\N	\N	25	94	28
143	2025-11-07 10:03:14.542609+00	2025-11-12 09:04:03.863466+00	\N	Jean	Tanguy	j.tanguy.22@gmail.com	\N	confirmed	2025-11-07 10:03:14.542417+00	2025-11-12 09:04:03.861542+00	2025-11-14 08:00:00+00	\N	\N	25	95	28
150	2025-11-09 09:53:55.225111+00	2025-11-09 09:53:55.225111+00	\N	Maéva	Cadeau	maeva.cadeau@orange.fr	06 19 98 01 99	pending	2025-11-09 09:53:55.224732+00	\N	\N	\N	\N	26	45	\N
155	2025-11-18 14:08:46.079626+00	2025-11-18 14:08:46.079626+00	\N	Valentin	Legendre	valentinvllegendre@gmail.com	\N	pending	2025-11-18 14:08:46.079244+00	\N	\N	\N	\N	26	12	\N
171	2026-05-08 15:44:37.445067+00	2026-05-08 15:44:37.445067+00	\N	Catherine	Jakubiec	catherinejaku@hotmail.com	0663497345	confirmed	2026-05-08 15:44:37.444894+00	2026-05-08 15:44:37.444894+00	\N	\N	\N	31	111	\N
172	2026-05-08 15:45:36.751157+00	2026-05-08 15:45:36.751157+00	\N	Zoé	CAREMEL	zoe.caremel@mailo.com	\N	confirmed	2026-05-08 15:45:36.750873+00	2026-05-08 15:45:36.750872+00	\N	\N	\N	31	112	\N
166	2026-04-28 09:17:19.086246+00	2026-04-29 15:47:48.054341+00	\N	Victor	DENIS	victordenis01@gmail.com	0652809335	confirmed	2026-04-28 09:17:19.085898+00	2026-04-29 15:47:48.052309+00	2026-05-01 07:00:00+00	\N	\N	30	62	\N
145	2025-11-07 21:15:50.096994+00	2025-11-23 10:10:53.497731+00	\N	Maéva	Cadeau	maeva.cadeau@orange.fr	06 19 98 01 99	confirmed	2025-11-07 21:15:50.096926+00	2025-11-23 10:10:53.497312+00	2025-11-25 13:30:00+00	\N	\N	22	45	\N
157	2026-04-14 18:07:27.490136+00	2026-04-22 16:57:29.659188+00	\N	Alain 	viel 	vieldelangon@wanadoo.fr	\N	confirmed	2026-04-14 18:07:27.489905+00	2026-04-22 16:57:29.656302+00	2026-04-24 08:00:00+00	\N	\N	29	106	\N
156	2026-04-14 18:04:25.215276+00	2026-04-22 17:00:52.722368+00	\N	anne	viel	vielderennes@orange.fr	\N	confirmed	2026-04-14 18:04:25.215209+00	2026-04-22 17:00:52.720386+00	2026-04-24 08:00:00+00	\N	\N	29	105	\N
160	2026-04-18 17:05:18.258116+00	2026-04-22 19:52:12.471191+00	\N	Maéva	Cadeau	maeva.cadeau@orange.fr	0619980199	confirmed	2026-04-18 17:05:18.258031+00	2026-04-22 19:52:12.469002+00	2026-04-24 08:00:00+00	\N	\N	29	45	32
161	2026-04-18 17:05:18.25927+00	2026-04-22 19:52:24.413409+00	\N	Lucas	Fournier	maeva.cadeau@orange.fr	\N	confirmed	2026-04-18 17:05:18.259222+00	2026-04-22 19:52:24.41151+00	2026-04-24 08:00:00+00	\N	\N	29	45	32
146	2025-11-08 14:32:37.796351+00	2025-11-26 05:48:23.329616+00	\N	Virginie	Vincent	vincent.virginie@gmail.com	0688728475	confirmed	2025-11-08 14:32:37.796181+00	2025-11-26 05:48:23.327614+00	2025-11-28 09:30:00+00	\N	\N	23	97	29
141	2025-10-31 09:45:53.924325+00	2025-11-26 09:07:18.633401+00	\N	Emilie	Massard	emiliemassard@wanadoo.fr	\N	confirmed	2025-10-31 09:45:53.924277+00	2025-11-26 09:07:18.631236+00	2025-11-28 13:30:00+00	\N	\N	24	93	\N
147	2025-11-08 14:32:37.800113+00	2025-11-26 15:37:49.314568+00	\N	Mélanie	lepretre	contact@papi-jean.com	\N	confirmed	2025-11-08 14:32:37.799923+00	2025-11-26 15:37:49.314222+00	2025-11-28 09:30:00+00	\N	\N	23	98	29
152	2025-11-09 20:56:04.109698+00	2025-11-27 10:23:31.915966+00	\N	Sophie	Lemarié 	y.lemarie@sfr.fr	0664215771	cancelled	2025-11-09 20:56:04.109637+00	\N	2025-11-28 09:30:00+00	2025-11-27 10:23:31.914546+00	No confirmation before deadline	23	102	\N
154	2025-11-14 16:59:34.358018+00	2025-12-18 16:43:42.505325+00	\N	Theo	Clemenceau 	theo.clemenceau352@gmail.com	\N	confirmed	2025-11-14 16:59:34.357931+00	2025-12-18 16:43:42.504266+00	2025-11-22 09:00:00+00	\N	\N	21	104	31
162	2026-04-18 17:05:18.259893+00	2026-04-22 19:52:27.092373+00	\N	Sienna	Fournier	maeva.cadeau@orange.fr	\N	confirmed	2026-04-18 17:05:18.259845+00	2026-04-22 19:52:27.090277+00	2026-04-24 08:00:00+00	\N	\N	29	45	32
165	2026-04-24 11:41:10.506733+00	2026-05-06 14:09:38.060745+00	\N	Sienna	Fournier	maeva.cadeau@orange.fr	\N	confirmed	2026-04-24 11:41:10.506556+00	2026-05-06 14:09:38.05863+00	2026-05-08 07:00:00+00	\N	\N	31	45	33
164	2026-04-24 11:41:10.505021+00	2026-05-06 14:09:45.268102+00	\N	Lucas	Fournier	maeva.cadeau@orange.fr	\N	confirmed	2026-04-24 11:41:10.504841+00	2026-05-06 14:09:45.26633+00	2026-05-08 07:00:00+00	\N	\N	31	45	33
163	2026-04-24 11:41:10.502283+00	2026-05-06 14:09:51.412958+00	\N	Maéva	Cadeau	maeva.cadeau@orange.fr	0619980199	confirmed	2026-04-24 11:41:10.50209+00	2026-05-06 14:09:51.411231+00	2026-05-08 07:00:00+00	\N	\N	31	45	33
168	2026-05-05 19:37:11.319701+00	2026-05-07 07:52:40.499881+00	\N	Séverine 	DESILLE 	redcorvette2018@gmail.com	\N	cancelled	2026-05-05 19:37:11.319528+00	\N	2026-05-08 07:00:00+00	2026-05-07 07:52:40.498609+00	No confirmation before deadline	31	109	\N
167	2026-04-28 21:49:01.931209+00	2026-05-01 07:57:05.430051+00	\N	Dorine 	Fourgaut 	fourgautdorine@yahoo.fr	0659738640	confirmed	2026-04-28 21:49:01.930846+00	2026-05-01 07:57:05.427572+00	2026-05-01 07:00:00+00	\N	\N	30	108	\N
159	2026-04-14 18:16:07.379055+00	2026-05-01 07:57:05.434342+00	\N	alain	viel	vieldelangon@wanadoo.fr	\N	confirmed	2026-04-14 18:16:07.378731+00	2026-05-01 07:57:05.43316+00	2026-05-01 07:00:00+00	\N	\N	30	106	\N
158	2026-04-14 18:14:21.353057+00	2026-05-01 07:57:05.437657+00	\N	anne	viel	vielderennes@orange.fr	\N	confirmed	2026-04-14 18:14:21.352728+00	2026-05-01 07:57:05.436634+00	2026-05-01 07:00:00+00	\N	\N	30	105	\N
169	2026-05-07 18:13:01.320112+00	2026-05-07 18:13:01.320112+00	\N	Yann	Le Doare 	yann@linuxconsole.org	0651031144	confirmed	2026-05-07 18:13:01.319936+00	2026-05-07 18:13:01.314853+00	\N	\N	\N	31	110	34
170	2026-05-07 18:13:01.322143+00	2026-05-07 18:13:01.322143+00	\N	Emilie 	Fichou	yann@linuxconsole.org	\N	confirmed	2026-05-07 18:13:01.321956+00	2026-05-07 18:13:01.314853+00	\N	\N	\N	31	110	34
185	2026-06-17 08:11:50.259549+00	2026-06-17 08:21:08.640758+00	\N	ana	Rolland 	rollandj.anais@gmail.com	\N	confirmed	2026-06-17 08:11:50.259086+00	2026-06-17 08:11:50.259086+00	2026-06-19 07:00:00+00	\N	\N	37	125	\N
186	2026-06-17 08:12:31.670243+00	2026-06-17 08:21:09.322668+00	\N	Benoît 	Rolland 	breizhbrj@yahoo.fr	\N	confirmed	2026-06-17 08:12:31.669624+00	2026-06-17 08:12:31.669624+00	2026-06-19 07:00:00+00	\N	\N	37	126	\N
174	2026-05-18 16:02:53.553722+00	2026-06-03 07:05:40.237351+00	\N	Agathe	Séné	agathesene0@gmail.com	\N	confirmed	2026-05-18 16:02:53.553494+00	2026-06-03 07:05:40.233258+00	2026-06-05 07:00:00+00	\N	\N	33	114	\N
173	2026-05-18 16:00:04.851139+00	2026-06-03 07:08:55.283033+00	\N	Anaëlle	Morel	agathe.senedu56@gmail.com	0638863382	confirmed	2026-05-18 16:00:04.850886+00	2026-06-03 07:08:55.27967+00	2026-06-05 07:00:00+00	\N	\N	33	113	\N
176	2026-06-04 08:13:18.276365+00	2026-06-04 08:13:18.276365+00	\N	Janine	Ruffault	vlnb12311@gmail.com	0617018106	confirmed	2026-06-04 08:13:18.275235+00	2026-06-04 08:13:18.275235+00	\N	\N	\N	33	2	\N
175	2026-05-31 13:52:28.418684+00	2026-06-04 08:28:01.619806+00	\N	Claire 	lebouc 	claire.freezone@gmail.com	0607273277	confirmed	2026-05-31 13:52:28.418226+00	2026-06-04 08:28:01.614074+00	2026-06-05 07:00:00+00	\N	\N	33	117	\N
177	2026-06-04 08:50:01.82395+00	2026-06-04 08:50:01.82395+00	\N	Victor	DENIS	victordenis01@gmail.com	0652809335	confirmed	2026-06-04 08:50:01.823488+00	2026-06-04 08:50:01.823488+00	\N	\N	\N	33	62	\N
178	2026-06-05 10:24:45.528465+00	2026-06-05 10:24:45.528465+00	\N	Emilie	Bourgeois	emibouu@gmail.com	0647167599	confirmed	2026-06-05 10:24:45.527928+00	2026-06-05 10:24:45.527928+00	\N	\N	\N	33	119	\N
192	2026-06-18 08:00:32.374742+00	2026-06-18 08:00:32.374742+00	\N	Maéva	Cadeau	maeva.cadeau@orange.fr	0619980199	confirmed	2026-06-18 08:00:32.373891+00	2026-06-18 08:00:32.366476+00	\N	\N	\N	37	45	37
193	2026-06-18 08:00:32.38472+00	2026-06-18 08:00:32.38472+00	\N	Lucas	Fournier	maeva.cadeau@orange.fr	\N	confirmed	2026-06-18 08:00:32.383824+00	2026-06-18 08:00:32.366476+00	\N	\N	\N	37	45	37
194	2026-06-18 08:00:32.393993+00	2026-06-18 08:00:32.393993+00	\N	Sienna	Fournier	maeva.cadeau@orange.fr	\N	confirmed	2026-06-18 08:00:32.392916+00	2026-06-18 08:00:32.366476+00	\N	\N	\N	37	45	37
191	2026-06-17 20:20:32.097384+00	2026-06-17 20:20:32.097384+00	\N	CELINE 	GUERIN 	guerice1@gmail.com	0684159138	pending	2026-06-17 20:20:32.096883+00	\N	\N	\N	\N	36	129	\N
182	2026-06-15 13:37:53.354046+00	2026-06-18 07:21:08.640166+00	\N	Maéva 	Cadeau 	maeva.cadeau@orange.fr 	0619980199	cancelled	2026-06-15 13:37:53.352925+00	\N	2026-06-19 07:00:00+00	2026-06-18 07:21:08.638154+00	No confirmation before deadline	37	124	35
183	2026-06-15 13:37:53.366233+00	2026-06-18 07:21:08.650823+00	\N	Lucas 	Fournier 	maeva.cadeau@orange.fr 	\N	cancelled	2026-06-15 13:37:53.365509+00	\N	2026-06-19 07:00:00+00	2026-06-18 07:21:08.648686+00	No confirmation before deadline	37	124	35
184	2026-06-15 13:37:53.374671+00	2026-06-18 07:21:08.660287+00	\N	Sienna 	Fournier 	maeva.cadeau@orange.fr 	\N	cancelled	2026-06-15 13:37:53.374127+00	\N	2026-06-19 07:00:00+00	2026-06-18 07:21:08.657922+00	No confirmation before deadline	37	124	35
195	2026-06-18 20:02:31.568764+00	2026-06-18 20:02:31.568764+00	\N	Jade	Guérin 	Jade35.guerin@gmail.com	0684159138	pending	2026-06-18 20:02:31.568079+00	\N	\N	\N	\N	36	130	\N
180	2026-06-11 18:42:03.960046+00	2026-06-26 09:51:46.939946+00	\N	Virginie 	Chérel	amelaint@yahoo.fr	0670126368	confirmed	2026-06-11 18:42:03.95928+00	2026-06-26 09:51:37.368906+00	2026-06-26 07:00:00+00	\N	\N	34	121	\N
179	2026-06-07 20:49:26.608441+00	2026-06-26 13:19:11.365136+00	\N	Damien	Closier	Closier.damien@postel.bzh	0631788884	cancelled	2026-06-07 20:49:26.607898+00	\N	2026-06-26 07:00:00+00	2026-06-26 13:19:11.360467+00	changement de plan de dernière minute	34	120	\N
181	2026-06-11 20:54:03.386746+00	2026-06-26 09:51:46.955529+00	\N	Jacques 	tixier	jtixier@free.fr	0663492358	confirmed	2026-06-11 20:54:03.38617+00	2026-06-26 09:51:46.952089+00	2026-06-26 07:00:00+00	\N	\N	34	122	\N
196	2026-06-27 03:28:00.417821+00	2026-06-27 03:28:00.417821+00	\N	Yassine	Amar	2.yassine.amar@gmail.com	0673468216	confirmed	2026-06-27 03:28:00.417323+00	2026-06-27 03:28:00.417322+00	\N	\N	\N	34	131	\N
197	2026-07-03 11:03:08.877753+00	2026-07-03 11:03:08.877753+00	\N	Emilie	Bourgeois	emibouu@gmail.com	0647167599	confirmed	2026-07-03 11:03:08.877252+00	2026-07-03 11:03:08.877251+00	\N	\N	\N	35	119	\N
187	2026-06-17 08:37:50.389599+00	2026-07-02 11:45:37.230373+00	\N	alain	viel	vielderennes@orange.fr	0610692449	confirmed	2026-06-17 08:37:50.38885+00	2026-07-02 11:45:37.227231+00	2026-07-03 12:30:00+00	\N	\N	35	105	36
190	2026-06-17 12:01:57.359913+00	2026-07-01 23:01:56.769984+00	\N	Thomas	VOISIN	baladeeco.retold540@passmail.net	0669439869	confirmed	2026-06-17 12:01:57.359353+00	2026-07-01 23:01:56.76569+00	2026-07-03 12:30:00+00	\N	\N	35	128	\N
188	2026-06-17 08:37:50.399467+00	2026-07-02 11:45:22.850321+00	\N	anne	viel	vielderennes@orange.fr	\N	confirmed	2026-06-17 08:37:50.398956+00	2026-07-02 11:45:22.846001+00	2026-07-03 12:30:00+00	\N	\N	35	105	36
189	2026-06-17 12:00:30.841798+00	2026-07-01 07:25:05.444261+00	\N	Léa	RION	lea.rion310@gmail.com	\N	confirmed	2026-06-17 12:00:30.841237+00	2026-07-01 07:25:05.439866+00	2026-07-03 12:30:00+00	\N	\N	35	127	\N
\.


--
-- Data for Name: rambles; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.rambles (id, created_at, updated_at, deleted_at, title, description, type, date, location, meeting_point, max_participants, difficulty, estimated_duration, equipment_needed, prerequisites, cover_image, additional_documents_url, is_cancelled, cancellation_date, cancellation_reason, payment_guide_id, payment_enabled, payment_required) FROM stdin;
1	2025-09-07 10:50:17.451711+00	2025-09-24 09:12:58.147045+00	\N	Initiation à la cueillette des champignons	La saison des cèpes et des chanterelles démarre, il est temps d'apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement et limiter son impact sur l'environnement. Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-04 12:30:00+00	Forêt domaniale de Paimpont	Parking du Chêne des Hindrés	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages car le parking des Hindrés est très fréquenté. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	1.jpg	\N	f	\N	\N	\N	f	f
2	2025-09-11 15:09:52.883609+00	2025-09-24 09:13:44.324259+00	\N	Initiation à la cueillette des champignons	La saison des cèpes et des chanterelles démarre, il est temps d'apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement et limiter son impact sur l'environnement. Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-08 12:30:00+00	Forêt domaniale de Paimpont	Parking du Chêne des Hindrés	15	facile	2:30:00	Habillez vous avec des vêtements longs et adaptés à la météo. Si vous en possédez, prenez avec vous un panier à fond plat, un couteau à champignons, un guide d'identification des champignons, une loupe de poche.	Nous allons essayer de mettre en place des covoiturages car le parking des Hindrés est très fréquenté. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	2.jpg	\N	f	\N	\N	\N	f	f
15	2025-09-26 08:27:50.581827+00	2025-10-09 09:44:49.383768+00	\N	A l'écoute du brame du cerf	Nous irons à la rencontre du roi de la forêt : le cerf, cet animal majestueux et impressionnant, que nous pourrons entendre pendant son brame. C’est une expérience marquante, ainsi que l’occasion de répondre à toutes vos questions.	faune	2025-10-11 17:30:00+00	Forêt de Paimpont	Parking de l'Abbaye de Paimpont	11	facile	2:30:00	Vêtements chauds et confortables, une petite lampe de poche ou frontale.	Nous irons à la rencontre du roi de la forêt : le cerf, cet animal majestueux et impressionnant, que nous pourrons entendre pendant son brame. C’est une expérience marquante, ainsi que l’occasion de répondre à toutes vos questions.	15.jpg	\N	t	2025-10-09 09:44:49.380342+00	Après plusieurs tentatives d'écoute, nous devons nous résoudre à l'évidence, le roi de la forêt s'est tu. La période du brame n'est pas une science exacte, et nous ne pouvons que nous y plier. C'est pourquoi nous ne pouvons maintenir cette sortie pourtant si magique, sans son chanteur principal. Ce message vous parvient avec toutes nos excuses pour cette annonce, mais aussi avec un lot de consolation : le cerf n'est pas le seul intérêt de la forêt à la nuit tombée, aussi proposons-nous une "contre-sortie" le même soir, même lieu, même heure, afin de ne pas rester sur cette mauvaise nouvelle. Si vous avez tout de même envie de vous promener dans les bois, soyez les bienvenu-e-s.	\N	f	f
10	2025-09-25 15:51:05.548931+00	2025-10-27 08:16:52.762294+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-29 09:00:00+00	Forêt domaniale de Paimpont	Parking du Chêne des Hindrés	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages car le parking des Hindrés et très prisé. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	10.jpg	\N	f	\N	\N	2	t	f
17	2025-09-26 08:34:53.192431+00	2025-10-17 21:13:28.940079+00	\N	Les oiseaux migrateurs à Careil	L’automne et l’hiver sont les périodes parfaites en Bretagne pour observer les oiseaux migrateurs dits hivernants, utilisant des milieux comme les étangs et les lacs pour passer l’hiver avant de remonter vers le Nord ou retourner vers la mer. Ils y rencontrent alors les résidents annuels de la zone. Je vous propose de suivre ce défilé tout au long de ces saisons à l’étang de Careil, qui est un lieu d’intérêt très réputé pour l’avifaune.	oiseaux	2025-10-18 07:00:00+00	Etang de Careil	Espace naturel départemental du domaine de Careil	15	facile	2:00:00	Vêtements chauds et confortables, si vous en possédez ramenez des jumelles, voire un guide sur les oiseaux.	Pensez à covoiturer, nous pouvons vous aider pour cela. Trois jours avant la sortie, vous recevrez un email qui demande de confirmer votre participation à la sortie et vous donnera la possiblité de payer en ligne, mais vous pouvez aussi régler en espèces sur place.	17.jpg	\N	f	\N	\N	\N	f	f
5	2025-09-24 09:17:14.693685+00	2025-10-15 20:13:48.877059+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-20 12:30:00+00	Forêt domaniale de Rennes	Parking de l'étang des Maffrais	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages pour limiter notre impact sur l'environnement. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	5.jpg	\N	f	\N	\N	2	t	f
12	2025-09-25 15:55:19.999148+00	2025-10-27 08:16:01.382556+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-31 09:00:00+00	Forêt domaniale de Paimpont	Parking du Tombeau de Merlin	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages car le parking est plutôt petit. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	12.jpg	\N	f	\N	\N	2	t	f
7	2025-09-24 09:23:36.664403+00	2025-10-15 20:14:51.252685+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-22 12:30:00+00	Forêt domaniale de Rennes	Relais Nature de Mi-Forêt	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages pour limiter notre impact sur l'environnement. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	7.jpg	\N	f	\N	\N	2	t	f
9	2025-09-25 15:38:51.617917+00	2025-10-22 10:09:38.748271+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-27 13:30:00+00	Forêt domaniale de Paimpont	Parking du Tombeau de Merlin	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages car les places de parking sont limitées. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	9.jpg	\N	f	\N	\N	2	t	f
13	2025-09-25 16:40:54.485976+00	2025-11-06 15:22:29.049542+00	\N	Cueillette et cuisine des champignons et plantes sauvages	Venez cueillir vos champignons avec Vera et ensuite préparer un repas sauvage avec Isabelle ! Nous combinons une balade de cueillette, avec des explications sur les champignons et comment les ramasser sans danger, suivie par un atelier de cuisine sauvage dans l'atelier d'Isabelle.	champignons	2025-11-11 09:30:00+00	Pléchâtel le Châtelier	Jardin des Merveilles	10	facile	5:00:00	Pensez à mettre des chaussures qui ne craignent pas la boue et sont étanches.\nAprès la cuisine, nous partagerons le repas : si vous avez des intolérances ou allergies alimentaires, faites-le nous savoir ^^	Venez comme vous êtes ! Tout le matériel est fourni. Nous allons organiser un covoiturage en départ de Rennes, car le car BreizhGo ne circule pas sur des horaires convenables ce jour férié.	13.jpg	\N	f	\N	\N	\N	f	f
8	2025-09-24 09:26:14.064439+00	2025-10-15 20:15:54.671152+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-24 08:00:00+00	Forêt domaniale de Rennes	Parking de l'étang des Maffrais	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages pour limiter notre impact sur l'environnement. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	8.jpg	\N	f	\N	\N	2	t	f
20	2025-10-09 09:18:36.581034+00	2025-10-11 17:55:10.698027+00	\N	Une nuit en forêt	Avez-vous déjà été dans les bois la nuit ? Entendu le chant des chouettes ? Aperçu la lueur d'un vers luisant ? Surpris les cris d'une chauve-souris ? Ce n'est pas parce que le soleil s'est couché que le sous-bois s'est endormi. Alors partons à l'aventure et promenons-nous dans les bois, tant que le jour n'y est pas.	decouverte	2025-10-11 17:30:00+00	Forêt de Brocéliande	Parking de l'abbaye de Paimpont	15	facile	2:30:00	Vêtements chauds et confortables, chaussures adaptées, une petite lampe de poche ou frontale.	A partir du point de rendez-vous, nous partirons en forêt avec le moins de voitures possibles et reviendrons après la balade au parking de l'abbaye. Si vous nécessitez un covoiturage n'hésitez pas à nous transmettre l'information.  	20.jpg	\N	f	\N	\N	2	t	f
6	2025-09-24 09:20:45.626806+00	2025-10-15 20:14:35.160573+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-22 08:00:00+00	Forêt domaniale de Rennes	Relais Nature de Mi-Forêt	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages pour limiter notre impact sur l'environnement. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	6.jpg	\N	f	\N	\N	2	t	f
14	2025-09-26 07:36:14.532684+00	2025-10-09 09:44:10.458883+00	\N	A l'écoute du brame du cerf	Nous irons à la rencontre du roi de la forêt : le cerf, cet animal majestueux et impressionnant, que nous pourrons entendre pendant son brame. C’est une expérience marquante, ainsi que l’occasion de répondre à toutes vos questions.	faune	2025-10-10 17:30:00+00	Forêt de Paimpont	Parking de l'Abbaye de Paimpont	13	facile	2:30:00	Vêtements chauds et confortables, une petite lampe de poche ou frontale.	Nous partirons en forêt avec le moins de voitures possible et ensuite nous reviendrons tous de nouveau à Paimpont récupérer les voitures individuelles. Trois jours avant la sortie vous recevrez un email pour confirmer votre participation : en cas d'annulation vous serez prévenus par email.	14.jpg	\N	t	2025-10-09 09:44:10.458036+00	Après plusieurs tentatives d'écoute, nous devons nous résoudre à l'évidence, le roi de la forêt s'est tu. La période du brame n'est pas une science exacte, et nous ne pouvons que nous y plier. C'est pourquoi nous ne pouvons maintenir cette sortie pourtant si magique, sans son chanteur principal. Ce message vous parvient avec toutes nos excuses pour cette annonce, mais aussi avec un lot de consolation : le cerf n'est pas le seul intérêt de la forêt à la nuit tombée, aussi proposons-nous une "contre-sortie" le même soir, même lieu, même heure, afin de ne pas rester sur cette mauvaise nouvelle. Si vous avez tout de même envie de vous promener dans les bois, soyez les bienvenu-e-s.	\N	f	f
19	2025-10-09 09:13:16.679094+00	2025-10-11 17:55:01.351687+00	\N	Une nuit en forêt	Avez-vous déjà été dans les bois la nuit ? Entendu le chant des chouettes ? Aperçu la lueur d'un vers luisant ? Surpris les cris d'une chauve-souris ? Ce n'est pas parce que le soleil s'est couché que le sous-bois s'est endormi. Alors partons à l'aventure et promenons-nous dans les bois, tant que le jour n'y est pas.	decouverte	2025-10-09 17:30:00+00	Forêt de Brocéliande	Parking de l'abbaye de Paimpont	15	facile	2:30:00	Vêtements chauds et confortables, chaussures adaptées, une petite lampe de poche ou frontale.	A partir du point de rendez-vous, nous partirons en forêt avec le moins de voitures possibles et reviendrons après la balade au parking de l'abbaye. Si vous nécessitez un covoiturage n'hésitez pas à nous transmettre l'information.  	19.jpg	\N	f	\N	\N	\N	f	f
35	2026-05-06 10:23:14.411843+00	2026-05-25 14:51:19.641801+00	\N	Randonnée naturaliste en forêt de Brocéliande	Explorez le côté nature de la forêt de Brocéliande, sur une boucle de 5 km, avec Vera LORENZETTI et Val FORTINA, mycologues et botanistes qui vous révèleront de la richesse des espèces et des milieux. Au programme : arbres et plantes sauvages comestibles, médicinales et toxiques, champignons, oiseaux et usages historiques de la forêt !	decouverte	2026-07-04 12:30:00+00	Forêt domaniale de Paimpont	Parking du chêne des Hindrés	10	facile	3:00:00	Bonnes chaussures de marche, imperméables de préférence, vêtements longs (attention aux tiques !!) et adaptés à la météo, de l'eau et un petit goûter si vous voulez.	Les chemins sont assez accessibles, peu de dénivelé mais parfois le revêtement n'est pas terminé, merci de me prévenir si vous avez des difficultés particulières à marcher.	35.jpg	\N	f	\N	\N	2	t	f
18	2025-09-29 08:35:50.709147+00	2025-10-24 11:45:32.789037+00	\N	Incendies de forêt et champignons	Vous allez découvrir l'évolution des Landes Rennaises, une zone de lande et forêt qui a été touchée par les incendies de 2022. Elle a été objet d'une étude sur les champignons liés aux incendies, elle a vu des animaux blessés par le feu être réintroduits après soins, on y a fait des travaux d'élagage et de plantation pour relancer et améliorer l'écosystème en place, en allant vers une forêt plus naturelle et durable. Nous vous proposons de découvrir ce site spécial et en apprendre plus sur son histoire depuis les feux, entre colonisation naturelle, choix d'aménagement et réintroduction de la faune sauvage.	decouverte	2025-10-28 13:30:00+00	Campénéac	Les Landes Rennaises	15	facile	2:00:00	Des chaussures étanches car nous irons près d'une mare et le sol peut être trempé.	Le terrain est globalement plat, la zone de landes nécessite de marcher sur les affleurements rocheux mais cela ne demande pas une condition physique particulière. Trois jours avant la sortie vous recevrez un email pour confirmer votre participation et le lien de payement en ligne (optionnel).	18.jpg	\N	t	2025-10-24 11:45:32.785417+00	Maéva vous êtes la seule inscrite sur ce créneau, du coup je vais annuler cette date mais j'aimerais reproposer la sortie en novembre ou début décembre. Est-ce que vous pourriez m'envoyer un email pour connaître vos disponibilités et vous permettre de participer à la prochaine sortie sur ce sujet svp ? Merci, bonne journée et bon weekend ! Vera	2	t	f
11	2025-09-25 15:52:54.204997+00	2025-10-28 09:40:37.609541+00	\N	Initiation à la cueillette des champignons	Vous désirez apprendre à reconnaître les champignons comestibles, éviter les confusions avec les toxiques, connaître les bons gestes pour cueillir proprement tout en limitant votre impact sur l'environnement ? Je vous propose de vous accompagner dans cet apprentissage, en toute sécurité.	champignons	2025-10-29 13:30:00+00	Forêt domaniale de Paimpont	Parking du Chêne des Hindrés	15	facile	2:30:00	Vêtements longs et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons essayer de mettre en place des covoiturages car le parking des Hindrés et très prisé. Trois jours avant la sortie, vous recevrez un email pour confirmer votre inscription et payer en ligne (si vous le souhaitez, il est possible aussi de payer en espèces le jour J).	11.jpg	\N	t	2025-10-28 09:40:37.606598+00	Titouan on se voit demain à 10h sur le parking du chêne des Hindrés, vu qu'il n'y a toujours pas de nouveaux inscrits on va tous se concentrer sur le groupe du matin :) bonne journée et à demain !	2	t	f
21	2025-10-30 09:15:01.709233+00	2025-10-30 09:18:48.405892+00	\N	Initiation à la cueillette des champignons	Nous sommes à la fin de la saison, c'est la dernière occasion de s'initier à la cueillette des champignons : chanterelles en tube, trompettes des morts et pieds bleus... Je vous présenterai les espèces comestibles et vous apprendrai les bons gestes pour une cueillette sécure et respectueuse de l'environnement.	champignons	2025-11-23 09:00:00+00	Etang des Maffrais	Parking du Parcours Ecologique	15	facile	2:30:00	Vêtements longs, chauds (gants et bonnet) et adaptés à la météo. Si vous en possédez, un panier à fond plat, un couteau type opinel, un guide d'identification des champignons et une loupe de poche.	Nous allons mettre en place des covoiturages pour limiter le numéro de voitures. N'oubliez pas de confirmer votre participation trois jours avant le jour J, vous recevrez un email à ce sujet qui vous permettra aussi de régler en ligne si vous le souhaitez.	21.jpg	\N	f	\N	\N	2	t	f
24	2025-10-30 09:32:02.537472+00	2025-10-30 09:32:49.620249+00	\N	Découverte des champignons	Sortons en forêt pour découvrir la diversité des champignons ! Comestibles ou non, grands et petits, nous prendrons le temps d'observer et d'apprivoiser la grande diversité de formes, couleurs, odeurs et saveurs des champignons.	champignons	2025-11-29 13:30:00+00	Mi-Forêt	Parking du Relais Nature	15	facile	2:00:00	Vêtements longs, chauds (bonnet et gants) et adaptés à la météo. Si vous en possédez, un guide d'identification des champignons.	Nous allons mettre en place des covoiturages pour limiter le nombre de voitures. N'oubliez pas de confirmer votre participation suite à l'email que vous recevrez 3 jours avant le jour J, vous pourrez aussi payer en ligne si vous le souhaitez.	24.jpg	\N	f	\N	\N	2	t	f
23	2025-10-30 09:28:20.907787+00	2025-10-30 09:30:22.096012+00	\N	Découverte des champignons	Sortons en forêt pour découvrir la diversité des champignons ! Comestibles ou non, grands et petits, nous prendrons le temps d'observer et d'apprivoiser la grande diversité de formes, couleurs, odeurs et saveurs des champignons.	champignons	2025-11-29 09:30:00+00	Mi-Forêt	Parking du Relais Nature	15	facile	2:00:00	Vêtements longs, chauds (bonnet et gants) et adaptés à la météo. Si vous en possédez, un guide d'identification des champignons.	Nous allons mettre en place des covoiturages pour limiter le nombre de voitures. N'oubliez pas de confirmer votre participation suite à l'email que vous recevrez 3 jours avant le jour J, vous pourrez aussi payer en ligne si vous le souhaitez.	23.jpg	\N	f	\N	\N	2	t	f
22	2025-10-30 09:25:01.869145+00	2025-11-23 18:49:48.901808+00	\N	Découverte des champignons	Sortons en forêt pour découvrir la diversité des champignons ! Comestibles ou non, grands et petits, nous prendrons le temps d'observer et d'apprivoiser la grande diversité de formes, couleurs, odeurs et saveurs des champignons.	champignons	2025-11-26 13:30:00+00	Juteauderies en forêt de Rennes	Parking près du rond point de la Petite Lune	15	facile	2:00:00	Vêtements longs, chauds (bonnet et gants) et adaptés à la météo. Si vous en possédez, un guide d'identification des champignons.	Nous allons organiser des covoiturages pour limiter le nombre de voitures. Depuis Rennes, il est possible de prendre le bus 50 en départ de Viasilva et descendre à Juteauderies, puis marches 10 minutes jusqu'au point de rdv.	22.jpg	\N	t	2025-11-23 18:49:48.897719+00	Nombre d'inscrits insuffisant, c'est parti pour le plan b !	2	f	f
25	2025-11-02 12:40:45.279716+00	2025-11-02 12:40:45.295809+00	\N	Les oiseaux migrateurs	L’automne et l’hiver sont les périodes parfaites en Bretagne pour observer les oiseaux migrateurs dits hivernants, utilisant des milieux comme les étangs et les lacs pour passer l’hiver avant de remonter vers le Nord ou retourner vers la mer. Ils y rencontrent alors les résidents annuels de la zone. Je vous propose de suivre ce défilé tout au long de ces saisons à l’étang de Careil, qui est un lieu d’intérêt très réputé pour l’avifaune.	oiseaux	2025-11-15 08:00:00+00	Domaine de Careil	Parking du Domaine de Careil	15	facile	2:00:00	Vêtements chauds et confortables, si vous en possédez ramenez des jumelles ou appareil photo, voire un guide sur les oiseaux.	Pensez à covoiturer, nous pouvons vous aider pour cela. Trois jours avant la sortie, vous recevrez un email qui demande de confirmer votre participation à la sortie et vous donnera la possiblité de payer en ligne, mais vous pouvez aussi régler en espèces sur place.	25.jpg	\N	f	\N	\N	\N	f	f
27	2026-03-13 16:46:29.797607+00	2026-03-13 17:35:08.765829+00	\N	Balade EcoLogique dans l'univers de Jeanne Macaigne	Sortie gratuite, idéale pour les familles, pour découvrir la nature en ville et l'univers de l'autrice Jeanne Macaigne, qui est au centre des P'tits bouquineurs 2026 avec sa vision colorée et curieuse de la nature. Vous trouverez ses oeuvres à la bibliothèque de l'Antipode ! \nNous allons mettre en avant des thématiques qui lui sont chères et vous permettre d'approcher et explorer la nature avec les yeux d'une fourmi, questionner ce qu'on voit et découvrir la richesse qui nous entoure.\nVous pouvez vous présenter devant l'entrée de la bibliothèque de l'Antipode à partir de 10h45, ce n'est pas nécessaire de réserver ses places.	decouverte	2026-03-21 10:00:00+00	Antipode	Devant l'entrée de la bibliothèque de l'Antipode	20	facile	1:30:00	Des chaussures fermées et des vêtements adaptés à la météo, nous allons marcher un peu en dehors des chemins dans le bosquet.	Venez comme vous êtes :)	27.jpg	\N	f	\N	\N	\N	f	f
26	2025-11-02 12:50:52.066107+00	2025-11-18 14:10:41.972696+00	\N	Sortie nocturne amphibiens	Partons à la découverte des salamandres et autres amphibiens qui pointent le bout du nez dans les sous-bois une fois la nuit tombée. Habillez-vous chaudement et n’oubliez pas vos lampes-torches !	faune	2025-11-22 18:00:00+00	Bois de Soeuvre	Parking de l'espace naturel départemental du Bois de Soeuvres	15	facile	2:00:00	Prenez des vêtements chauds et imperméables, des chaussures qui ne craignent pas la boue. Une lampe-torche serait pertinente.	Nous organiserons des covoiturages pour limiter notre impact et optimiser nos déplacements. La balade sera susceptible d'être annulée en cas de météo trop clémente.	26.jpg	\N	t	2025-11-18 14:10:41.969228+00	Bonjour,\n  C'est avec regret que nous devons annuler la Sortie nocturne amphibiens. La Bretagne subit cette semaine une vague de froid, amenant du givre matinal. Nos amies salamandres vont devoir s'en protéger et partiront ainsi probablement hiberner d'ici quelques jours. La probabilité de croiser cet élégant amphibien devient donc faible, et représente le motif d'annulation de la balade. Souhaitons-leur un bon hiver, et donnons-leur rendez-vous au le printemps prochain, vers Mars.	\N	f	f
28	2026-03-13 16:58:37.800459+00	2026-03-13 17:38:18.496943+00	\N	Grand pique-nique de printemps	Rejoignez-nous au parc des Gayeulles pour fêter le début du printemps et la reprise des sorties de Balade EcoLogique ! Nous vous proposons de pique-niquer tous ensemble pour partager un moment convivial, se revoir (ou se rencontrer !!), papoter, parler des nouveautés et projets à venir, jouer...\nOn se retrouve à partir de midi sur la prairie en haut du petit étang à bateaux de modélisme, sur la gauche de l'accrobranche si vous rentrez par l'entrée à gauche de la piscine. \nPOINT GPS : https://maps.app.goo.gl/DoMAniemQ9F4ttgT7\n	decouverte	2026-03-22 11:00:00+00	Parc des Gayeulles	prairie en haut à gauche de l'accrobranche	100	facile	3:00:00	Amenez votre propre repas (et un petit quelque chose à partager ? nous vous partagerons aussi un petit quelque chose de spécial à goûter), une couverture de sol, et pourquoi pas un ballon ou un petit jeu sympa ^^	Amenez toute personne intéressée par le projet et qui a envie de partager un moment convivial avec nous tous :) plus on est de fous...	28.jpg	\N	f	\N	\N	\N	f	f
30	2026-03-23 19:08:25.412161+00	2026-04-24 10:29:54.161893+00	\N	Randonnée naturaliste en forêt de Rennes	Parcourez quelques kilomètres autour de l'Etang de Maffrais aux côtés de Vera, mycologue et botaniste qui vous parlera de la richesse des espèces et des milieux qu'on y rencontrera. Au menu : plantes sauvages comestibles, médicinales et toxiques, champignons de printemps, oiseaux et lichens !	decouverte	2026-05-02 07:00:00+00	Forêt de Rennes	Parking de l'étang des Maffrais	10	facile	4:00:00	Bonnes chaussures de marche qui ne craignent pas la boue, imperméables. Vêtements longs et adaptés à la météo. Une gourde d'eau, un snack en cas de petit creux. La randonnée de termine à 13h, vous pourrez enchaîner sur un pique-nique en forêt si vous le souhaitez !	Si vous avez des problèmes de santé ou à marcher, merci de nous en prévenir.	30.jpg	\N	f	\N	\N	2	t	f
31	2026-04-22 12:54:20.254253+00	2026-04-22 12:54:20.270052+00	\N	L'opéra des oiseaux aux Gayeulles	En ce moment les oiseaux nidifient, ce qui veut dire qu'ils doivent à la fois s'occuper de leur territoire contre les voisins ambitieux, et du nourrissage de leurs potentiels jeunes. Entre chant et ballet, venez observer l'opéra des Gayeulles sur une matinée.	oiseaux	2026-05-09 07:00:00+00	Parc des Gayeulles	Parc des Gayeulles Car Park	15	facile	2:00:00	Chaussures adaptées à la marche, si possible une paire de jumelles.		31.jpg	\N	f	\N	\N	\N	f	f
29	2026-03-23 18:56:06.246289+00	2026-04-06 10:28:12.633222+00	\N	Des champignons toute l'année !	Mais où sont les champignons au printemps ? Eh bien, partout aoutour de nous ! Je vous accompagne dans une découverte du monde incroyable des champignons, au delà des petites choses parfois très gourmandes que nous avons l'habitude de croiser à l'automne. Vous allez découvrir les champignons qui sont là toute l'année, sans qu'on les remarque, ainsi que les champignons printaniers (et il y en a aussi des comestibles !). Pour vous récompenser de votre curiosité, vous recevrez une vidéo animée et un document pédagogique suite à notre sortie.	decouverte	2026-04-25 08:00:00+00	Parc des Gayeulles	Devant l'entrée de la patinoire Le BLizz	15	facile	2:00:00	Des chaussures fermées pour marcher hors chemin, vêtements longs et adaptés à la météo.	Beaucoup de curiosité pour découvrir ce monde magnifique mais invisible aux yeux de ceux qui ne veulent pas voir..!	29.jpg	\N	f	\N	\N	2	t	f
32	2026-04-22 14:39:51.780786+00	2026-05-08 11:10:33.429811+00	\N	Randonnée naturaliste en forêt de Brocéliande	Explorez le côté nature de la forêt de Brocéliande, sur une boucle de 5 km, avec Vera LORENZETTI et Val FORTINA, mycologues et botanistes qui vous révèleront de la richesse des espèces et des milieux. Au programme : arbres et plantes sauvages comestibles, médicinales et toxiques, champignons, oiseaux et usages historiques de la forêt !	decouverte	2026-05-09 12:30:00+00	Forêt domaniale de Paimpont	Parking du chêne des Hindrés	10	facile	3:00:00	Bonnes chaussures de marche, imperméables de préférence, vêtements longs (attention aux tiques !!) et adaptés à la météo, de l'eau et un petit goûter si vous voulez.	Les chemins sont assez accessibles, peu de dénivelé mais parfois le revêtement des fois n'est pas terminé, merci de me prévenir si vous avez des difficultés particulières à marcher.	32.jpg	\N	t	2026-05-08 11:10:33.42743+00	pas d'inscrits :)	2	t	f
33	2026-05-06 10:13:43.662826+00	2026-05-20 09:49:30.408495+00	\N	Randonnée naturaliste de découverte de la forêt de Rennes	Aventurez-vous sur le sentier pédestre de Caleuvre, entre hêtres, pins et chênes centenaires, à la découverte de la nature qui vous entoure. Nous traverserons plusieurs milieux pour observer arbres, plantes comestibles et médicinales, oiseaux, champignons et lichens, sur un parcours d'environ 7 km. Ce sera aussi l'occasion de parler d'écologie de manière scientifique et mieux comprendre les multiples relations qui nous relient avec les autres êtres vivants, dans un regard curieux et ouvert sur la nature.\nAprès la rando (vers 13h) vous pourrez pique-niquer en forêt, il y a des tables en bois à disposition aux Juteauderies.	decouverte	2026-06-06 07:00:00+00	Forêt domaniale de Rennes	Parking des Juteauderies	10	facile	4:00:00	Chaussures de marche confortables et étanches, vêtements longs (attention aux tiques !) conformes à la météo, une gourde d'eau. Vos guides naturalistes de terrain si l'envie y est :) et un pique-nique si vous le souhaitez.	Le chemin n'est pas particulièrement difficile, merci de nous informer si vous avez des difficultés particulières à marcher ou des problèmes de santé.	33.jpg	\N	f	\N	\N	2	t	f
34	2026-05-06 10:16:41.390226+00	2026-05-20 10:23:15.546264+00	\N	Randonnée naturaliste de découverte de la forêt de Rennes	Aventurez-vous sur le sentier pédestre de Caleuvre, entre hêtres, pins et chênes centenaires, à la découverte de la nature qui vous entoure. Nous traverserons plusieurs milieux pour observer arbres, plantes comestibles et médicinales, oiseaux, champignons et lichens, sur un parcours d'environ 7 km. Ce sera aussi l'occasion de parler d'écologie de manière scientifique et mieux comprendre les multiples relations qui nous relient avec les autres êtres vivants, dans un regard curieux et ouvert sur la nature.\nAprès la rando (vers 13h) vous pourrez pique-niquer en forêt, il y a des tables en bois à disposition aux Juteauderies.	decouverte	2026-06-27 07:00:00+00	Forêt domaniale de Rennes	Parking des Juteauderies	10	facile	4:00:00	Chaussures de marche confortables et étanches, vêtements longs (attention aux tiques !) conformes à la météo, une gourde d'eau. Vos guides naturalistes de terrain si l'envie y est :) et un pique-nique si vous le souhaitez.	Le chemin n'est pas particulièrement difficile, merci de nous informer si vous avez des difficultés particulières à marcher ou des problèmes de santé.	34.jpg	\N	f	\N	\N	2	t	f
37	2026-06-04 08:25:05.738454+00	2026-06-15 13:00:43.823249+00	\N	L'opéra des oiseaux aux Gayeulles	Les oiseaux continuent leur nidification en ce début d'été, ce qui veut dire qu'ils doivent à la fois s'occuper de leur territoire contre les voisins ambitieux, et du nourrissage de leurs jeunes.\n\nEntre chants et aller-retours frénétiques, venez observer l'opéra des Gayeulles durant une matinée.	oiseaux	2026-06-20 07:00:00+00	Parc des Gayeulles	Parking Piscine des Gayeulles	15	facile	2:00:00	Chaussures adaptées à la marche, si possible une paire de jumelles.	Aucune	37.jpg	\N	f	\N	\N	\N	f	f
36	2026-05-06 13:58:22.84243+00	2026-06-12 15:40:56.091883+00	\N	Randonnée naturaliste en forêt de Brocéliande	Explorez le côté nature de la forêt de Brocéliande, sur une boucle de 5 km, avec Val FORTINA, mycologue et botaniste, et Arnaud Crèvecoeur, artiste et botaniste, qui vous révèleront de la richesse des espèces et des milieux. Au programme : arbres et plantes sauvages comestibles, médicinales et toxiques, champignons, oiseaux et usages historiques de la forêt !	decouverte	2026-07-11 12:30:00+00	Forêt domaniale de Paimpont	Parking du chêne des Hindrés	10	facile	3:00:00	Bonnes chaussures de marche, imperméables de préférence, vêtements longs (attention aux tiques !!) et adaptés à la météo, de l'eau et un petit goûter si vous voulez.	Les chemins sont assez accessibles, peu de dénivelé mais parfois le revêtement des fois n'est pas terminé, merci de me prévenir si vous avez des difficultés particulières à marcher.	36.jpg	\N	f	\N	\N	\N	f	f
40	2026-06-12 15:16:31.889999+00	2026-06-15 12:46:40.15003+00	\N	Découverte sensorielle de la nature au crépuscule	Quand la lumière baisse, et que les chiens pourraient passer pour des loups dans la pénombre, venez découvrir en tout petit groupe une biodiversité (extra)ordinaire avec l'ensemble de vos 5 sens !\nLors d'une balade d'environ 5km, nous traverserons différents milieux (zone humide, bocage, prairie) pour apprendre à lire le paysage, reconnaître des végétaux, connaître leurs usages traditionnels, sentir la bonne odeur des fleurs du soir, observer les animaux qui s'activent à la tombée du jour. Peut-être aurons nous la chance d'entendre une chouette, croiser des vers luisants ou voir quelques chauves-souris... 	decouverte	2026-06-28 18:30:00+00	Concoret	parking de la salle éon de l'étoile	8	facile	1:30:00	Vêtements adaptés à la météo, chaussures fermées. Si vous avez une lampe frontale, vous pouvez la prendre.	Gratuit et sans inscription pour les enfants de moins de 6 ans (accompagnés par un adulte).	40.jpg	\N	f	\N	\N	\N	f	f
38	2026-06-12 14:29:01.374896+00	2026-06-15 13:46:49.75892+00	\N	Entre chiens et loups	Quand la lumière baisse, et que les chiens pourraient passer pour des loups dans la pénombre, venez découvrir en tout petit groupe une biodiversité (extra)ordinaire avec l'ensemble de vos 5 sens !\nLors d'une balade d'environ 5km, nous traverserons différents milieux (zone humide, bocage, prairie), pour apprendre à lire le paysage, reconnaître des végétaux, connaître leurs usages traditionnels, sentir la bonne odeur des fleurs du soir, observer les animaux qui s'activent à la tombée du jour. Peut-être aurons nous la chance d'entendre une chouette, croiser des vers luisants ou voir quelques chauves-souris... 	decouverte	2026-06-27 18:30:00+00	Concoret	parking de la salle éon de l'étoile, 56430 CONCORET	8	facile	1:30:00	Vêtements adaptés à la météo, chaussures fermées. Si vous avez une loupe frontale, vous pouvez la prendre.	Gratuit et sans inscription pour les enfants de moins de 6 ans (accompagnés par un adulte).	38.jpg	\N	f	\N	\N	\N	f	f
39	2026-06-12 14:49:01.141164+00	2026-06-23 13:49:27.173506+00	\N	Balade au crépuscule	Quand la lumière baisse, et que les chiens pourraient passer pour des loups dans la pénombre, venez découvrir en tout petit groupe une biodiversité (extra)ordinaire avec l'ensemble de vos 5 sens !\nLors d'une balade d'environ 5km, nous traverserons différents milieux (zone humide, bocage, prairie) pour apprendre à lire le paysage, reconnaître des végétaux, connaître leurs usages traditionnels, sentir la bonne odeur des fleurs du soir, observer les animaux qui s'activent à la tombée du jour. Peut-être aurons nous la chance d'entendre une chouette, croiser des vers luisants ou voir quelques chauves-souris... 	decouverte	2026-06-24 18:30:00+00	Concoret	parking de la salle éon de l'étoile	8	facile	1:30:00	Vêtements adaptés pour la météo. Si vous avez une lampe frontale, vous pouvez l'amener.	Gratuit et sans inscription pour les enfants de moins de 6 ans (accompagnés par un adulte).	39.jpg	\N	t	2026-06-23 13:49:27.161279+00	Annulée à cause de la canicule, d'autres dates sont ouvertes !	\N	f	f
41	2026-07-07 07:58:23.318927+00	2026-07-07 08:05:07.971664+00	\N	Balade au crépuscule	Quand la lumière baisse, et que les chiens pourraient passer pour des loups dans la pénombre, venez découvrir en tout petit groupe une biodiversité (extra)ordinaire avec l'ensemble de vos 5 sens !\nLors d'une balade d'environ 5km, nous traverserons différents milieux (zone humide, bocage, prairie) pour apprendre à lire le paysage, reconnaître des végétaux, connaître leurs usages traditionnels, sentir la bonne odeur des fleurs du soir, observer les animaux qui s'activent à la tombée du jour. Peut-être aurons nous la chance d'entendre une chouette, croiser des vers luisants ou voir quelques chauves-souris... 	decouverte	2026-07-14 18:30:00+00	Concoret	parking de la salle éon de l'étoile	8	facile	1:30:00	Vêtements adaptés à la météo, chaussures fermées. Si vous avez une lampe frontale, vous pouvez la prendre.	Gratuit et sans inscription pour les enfants de moins de 6 ans accompagnés par un adulte. Chiens tenus en laisse acceptés, prévenez-moi si vous amenez votre toutou :)	41.jpg	\N	f	\N	\N	\N	f	f
42	2026-07-07 08:41:36.787306+00	2026-07-07 08:42:38.148815+00	\N	Entre chiens et loups	sd	decouverte	2026-07-13 18:30:00+00	Paimpont	Parking P1 (rue des forges), au fond vers la table de pique-nique	8	facile	1:30:00			42.jpg	\N	f	\N	\N	\N	f	f
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.role_permissions (role_id, permission_id) FROM stdin;
1	1
1	2
1	3
1	4
1	5
1	6
1	7
2	3
2	5
1	8
1	9
1	10
1	11
1	12
1	13
1	14
2	9
2	10
2	12
1	15
1	16
1	17
1	18
1	19
1	20
1	21
1	22
1	23
1	24
1	25
3	9
3	10
3	8
3	12
3	14
3	3
3	5
3	26
3	27
3	28
3	29
3	30
3	31
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.roles (id, created_at, updated_at, deleted_at, name) FROM stdin;
1	2025-09-06 17:09:51.257125+00	2025-10-11 17:20:19.234563+00	\N	admin
3	2026-05-25 14:08:06.629979+00	2026-05-25 14:34:32.493056+00	\N	guide
2	2025-09-06 17:09:51.25873+00	2026-06-12 13:54:45.827894+00	\N	
\.


--
-- Data for Name: seeds; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.seeds (id, created_at, updated_at, deleted_at, name) FROM stdin;
1	2025-09-06 17:09:51.259498+00	2025-09-06 17:09:51.259498+00	\N	SeedV1
2	2025-09-06 17:09:51.265003+00	2025-09-06 17:09:51.265003+00	\N	SeedV2
3	2025-10-05 11:16:15.63185+00	2025-10-05 11:16:15.63185+00	\N	SeedV3
4	2025-10-05 11:16:15.635563+00	2025-10-05 11:16:15.635563+00	\N	SeedV4
5	2025-10-11 17:20:19.236179+00	2025-10-11 17:20:19.236179+00	\N	SeedV5
6	2026-05-25 14:08:06.631369+00	2026-05-25 14:08:06.631369+00	\N	SeedV6
7	2026-05-25 14:08:06.635063+00	2026-05-25 14:08:06.635063+00	\N	SeedV7
8	2026-07-02 14:56:07.527159+00	2026-07-02 14:56:07.527159+00	\N	SeedV8
\.


--
-- Data for Name: user_permission_overrides; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.user_permission_overrides (id, created_at, updated_at, deleted_at, user_id, permission_id, is_granted) FROM stdin;
\.


--
-- Data for Name: user_profiles; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.user_profiles (id, created_at, updated_at, deleted_at, first_name, last_name, avatar_name, phone, address_id, user_id) FROM stdin;
32	2025-10-02 13:49:50.157585+00	2025-10-07 08:12:36.102943+00	\N	jocelyne	giraux	\N	0685423249	\N	32
2	2025-09-11 15:11:47.346093+00	2025-09-11 15:11:47.346093+00	\N	vera	lor	\N	\N	\N	2
4	2025-09-25 17:32:54.818592+00	2025-09-25 17:32:54.818592+00	\N	Isabelle 	Cheval 	\N	\N	\N	4
44	2025-10-03 19:31:11.686756+00	2025-10-07 08:36:09.199862+00	\N	Pauline	Trion	\N	0633867441	\N	44
58	2025-10-07 10:18:06.794234+00	2025-10-07 10:18:06.794234+00	\N	Lenny 	Legrand 	\N	0671187096	\N	58
6	2025-09-28 18:15:07.372328+00	2025-09-28 18:15:07.372328+00	\N	hugo	Chapel 	\N	\N	\N	6
10	2025-09-30 11:18:17.151749+00	2025-09-30 11:18:17.151749+00	\N	Julien 	TRUBAT 	\N	0625927852	\N	10
11	2025-09-30 11:18:17.155419+00	2025-09-30 11:18:17.155419+00	\N	Perrine 	Oltra 	\N	\N	\N	11
12	2025-09-30 21:06:25.73803+00	2025-09-30 21:06:46.904749+00	\N	Valentin	Legendre	\N	0622994797	\N	12
13	2025-09-30 21:45:49.626717+00	2025-09-30 21:45:49.626717+00	\N	stefania	porcelli	\N	0769558635	\N	13
14	2025-10-01 04:21:56.776499+00	2025-10-01 04:21:56.776499+00	\N	anne	marchand	\N	0681265968	\N	14
18	2025-10-01 09:46:43.610001+00	2025-10-01 09:46:43.610001+00	\N	Sacha	Pérocheau Arnaud 	\N	\N	\N	18
19	2025-10-01 09:47:42.879008+00	2025-10-01 09:47:42.879008+00	\N	Françoise 	RAZANAMARO	\N	0688103431	\N	19
20	2025-10-01 09:47:42.881793+00	2025-10-01 09:47:42.881793+00	\N	Damien	BARBEDETTE	\N	\N	\N	20
21	2025-10-01 10:28:52.251426+00	2025-10-01 10:28:52.251426+00	\N	Philippe	cuillerier	\N	0675304855	\N	21
22	2025-10-01 10:31:10.588958+00	2025-10-01 10:31:42.824094+00	\N	xavier	even-cuilerier	\N	\N	\N	22
16	2025-10-01 06:56:19.033178+00	2025-10-01 11:06:55.777853+00	\N	Jean-Pierre	Guillot	\N	0633740654	\N	16
23	2025-10-01 16:25:00.232895+00	2025-10-01 16:25:00.232895+00	\N	Julie	ANBERREE	\N	\N	\N	23
24	2025-10-01 16:25:00.235523+00	2025-10-01 16:25:00.235523+00	\N	Thomas	ANBERREE	\N	\N	\N	24
26	2025-10-01 19:03:30.205539+00	2025-10-01 19:03:30.205539+00	\N	Delphine	BECHETOILLE	\N	\N	\N	26
27	2025-10-01 19:04:54.606396+00	2025-10-01 19:04:54.606396+00	\N	Charles	KERSUAL	\N	06 23 20 70 54	\N	27
34	2025-10-03 08:20:04.850882+00	2025-10-03 08:20:04.850882+00	\N	Marion 	Bezier 	\N	0634983715	\N	34
35	2025-10-03 08:20:04.853846+00	2025-10-03 08:20:04.853846+00	\N	Quentin 	Bellamy 	\N	\N	\N	35
36	2025-10-03 11:13:10.857943+00	2025-10-03 11:13:10.857943+00	\N	Anne-Marie	Guillot	\N	\N	\N	36
38	2025-10-03 12:08:10.600736+00	2025-10-03 12:08:10.600736+00	\N	Rébecca	GUYOT	\N	\N	\N	38
42	2025-10-03 16:46:20.554244+00	2025-10-03 16:46:20.554244+00	\N	stephanie	le labousse	\N	0681781953	\N	42
43	2025-10-03 16:47:24.20399+00	2025-10-03 16:47:24.20399+00	\N	christophe	berthault	\N	\N	\N	43
54	2025-10-06 16:04:32.486508+00	2025-10-28 05:44:46.531368+00	\N	adeline	Léon 	\N	\N	\N	54
39	2025-10-03 12:08:10.603729+00	2025-10-08 05:52:34.206033+00	\N	Julie	Glemot	\N	\N	\N	39
7	2025-09-29 06:17:58.559734+00	2025-10-05 06:21:44.22731+00	\N	Soizic	POILVE	\N	0634513252	\N	7
48	2025-10-05 10:59:30.060228+00	2025-10-05 10:59:30.060228+00	\N	Erwan 	Viéville 	\N	\N	\N	48
49	2025-10-05 10:59:30.062941+00	2025-10-05 10:59:30.062941+00	\N	Rosalba	Viéville 	\N	\N	\N	49
50	2025-10-05 10:59:30.064427+00	2025-10-05 10:59:30.064427+00	\N	Yvon	Viéville 	\N	\N	\N	50
17	2025-10-01 09:46:43.606981+00	2025-10-08 06:53:50.746313+00	\N	Clotilde 	Philippe 	\N	0667339738	\N	17
52	2025-10-05 14:42:12.968108+00	2025-10-05 14:42:12.968108+00	\N	Canelle 	Thomas	\N	0667087378	\N	52
53	2025-10-05 14:42:12.96964+00	2025-10-05 14:42:12.96964+00	\N	Charlotte 	Rondel	\N	\N	\N	53
28	2025-10-01 19:28:56.249086+00	2025-10-15 07:22:55.240552+00	\N	jean yves	dabo	\N	0643802854	\N	28
47	2025-10-04 19:25:30.4736+00	2025-10-15 06:36:19.184696+00	\N	mickael	dardaillon	\N	\N	\N	47
37	2025-10-03 12:01:54.385786+00	2025-10-08 08:11:21.534539+00	\N	Edith	BLIN	\N	06 84 17 81 81	\N	37
74	2025-10-16 09:06:39.680963+00	2025-10-16 09:07:16.882727+00	\N	edith	\N	\N	\N	\N	74
62	2025-10-11 17:40:12.392088+00	2026-05-25 14:24:09.990291+00	\N	Victor	DENIS	\N	\N	\N	62
61	2025-10-10 09:01:47.0307+00	2025-10-10 09:01:47.0307+00	\N	Soizic	MASSON	\N	07 83 33 53 52	\N	61
51	2025-10-05 14:27:26.390818+00	2025-10-17 05:06:34.255271+00	\N	Renaud	Delannay	\N	0608151402	\N	51
70	2025-10-13 22:02:19.166069+00	2025-10-15 11:50:22.412683+00	\N	Apolline	Privat	\N	\N	\N	70
46	2025-10-04 19:25:30.469942+00	2025-10-15 06:41:07.942616+00	\N	gaelle	tworkowski	\N	0673009969	\N	46
29	2025-10-01 19:28:56.251677+00	2025-10-15 12:26:48.048568+00	\N	nathalie	muller	\N	0687340223	\N	29
60	2025-10-09 18:33:00.146772+00	2025-10-19 06:15:14.925718+00	\N	cyril	pinchon	\N	0634457421	\N	60
55	2025-10-07 06:57:29.463157+00	2025-10-17 06:11:25.265506+00	\N	Nadège 	Lécrivain	\N	0674659708	\N	55
84	2025-10-23 11:36:24.952191+00	2025-10-23 11:36:24.952191+00	\N	mathilde	LEFRERE	\N	\N	\N	84
41	2025-10-03 15:53:36.489577+00	2025-11-09 15:32:49.750699+00	\N	fabienne	balcon	\N	0609948591	\N	41
31	2025-10-02 12:49:19.729668+00	2025-10-19 13:55:56.231039+00	\N	Isabelle	Lins	\N	0771122770	\N	31
30	2025-10-02 08:19:41.270683+00	2025-10-19 07:22:37.818422+00	\N	Mathieu	BELLEC	\N	0750038615	\N	30
65	2025-10-12 10:05:10.140199+00	2025-10-22 06:41:39.021811+00	\N	Lara	Schembri	\N	0682104561	\N	65
66	2025-10-13 17:22:50.710834+00	2025-10-19 17:42:33.071538+00	\N	Laura 	Giommi 	\N	0678461720	\N	66
67	2025-10-13 17:24:48.720867+00	2025-10-19 17:44:33.81884+00	\N	Solène 	Barbé 	\N	\N	\N	67
78	2025-10-20 08:11:18.575+00	2025-10-20 08:11:18.575+00	\N	Malika	TENEUR	\N	\N	\N	78
77	2025-10-19 17:52:18.683656+00	2025-10-19 17:57:31.384435+00	\N	manon	demeulenaere	\N	\N	\N	77
64	2025-10-12 10:05:10.134845+00	2025-10-20 08:50:01.567248+00	\N	Nina	Feurgard	\N	\N	\N	64
57	2025-10-07 08:15:51.328576+00	2025-10-20 21:04:32.631049+00	\N	pascal	Navarre	\N	0680069086	\N	57
80	2025-10-21 11:32:11.434782+00	2025-10-21 11:32:11.434782+00	\N	estelle	clua	\N	0604018719	\N	80
56	2025-10-07 08:12:46.867417+00	2025-10-20 21:05:57.620055+00	\N	martine	piel	\N	0681642064	\N	56
15	2025-10-01 05:25:04.074535+00	2025-10-21 07:43:04.71098+00	\N	Yannick 	Robert 	\N	0679821689	\N	15
79	2025-10-20 10:03:27.817149+00	2025-10-21 07:46:15.593082+00	\N	david	lecocq	\N	0612335399	\N	79
81	2025-10-21 11:57:00.25213+00	2025-10-21 11:57:00.25213+00	\N	diane 	Schmidt 	\N	\N	\N	81
83	2025-10-22 09:48:58.617019+00	2025-10-22 09:48:58.617019+00	\N	Béatricen	Rabault	\N	0676057291	\N	83
59	2025-10-07 14:06:22.260083+00	2025-10-21 15:32:34.17305+00	\N	iven	Le Louedec	\N	0649810682	\N	59
40	2025-10-03 15:13:00.800493+00	2025-10-22 09:50:26.780267+00	\N	Béatrice 	Rabault	\N	0676057291	\N	40
68	2025-10-13 18:20:09.173207+00	2025-10-22 11:34:06.412528+00	\N	Francis	DAHERON	\N	0652619988	\N	68
8	2025-09-29 14:46:53.06319+00	2025-10-24 05:03:35.461936+00	\N	Lolita	COZETTE 	\N	0640395643	\N	8
73	2025-10-15 16:49:09.777256+00	2025-10-24 16:44:38.188835+00	\N	David 	Octau 	\N	0629455581	\N	73
69	2025-10-13 22:00:05.962289+00	2025-10-24 09:56:06.588861+00	\N	Gaël Yann 	Rubin 	\N	0626877254	\N	69
33	2025-10-02 18:27:29.733782+00	2025-10-24 16:48:28.984221+00	\N	Elodie	Octau	\N	0672340773	\N	33
71	2025-10-14 19:48:53.314422+00	2025-10-29 19:10:02.149361+00	\N	Alexia	Muzas	\N	0625358869	\N	71
25	2025-10-01 19:03:30.202983+00	2025-10-26 09:07:31.097542+00	\N	Jeanne	LICHOU	\N	0623682355	\N	25
82	2025-10-21 12:09:25.375048+00	2025-10-26 11:43:30.712967+00	\N	Titouan	Millon	\N	0650040906	\N	82
5	2025-09-28 18:15:07.368746+00	2025-10-28 07:08:18.710089+00	\N	coline	vandepeutte	\N	0683619333	\N	5
63	2025-10-12 09:22:56.360541+00	2025-10-29 12:53:51.585589+00	\N	Romain	Prevosteau 	\N	0628451289	\N	63
75	2025-10-17 16:43:42.969552+00	2025-10-29 19:09:14.084085+00	\N	Amandine	Jegat	\N	0625358869	\N	75
9	2025-09-29 21:11:56.237358+00	2025-11-08 06:26:32.332088+00	\N	Marie Pierre	REMEUR 	\N	\N	\N	9
3	2025-09-25 16:46:09.74752+00	2025-11-14 07:38:24.79682+00	\N	Vera	Lorenzetti	\N	\N	\N	3
76	2025-10-18 09:48:07.789523+00	2025-10-24 04:58:58.970185+00	\N	Anthony 	Ambroise 	\N	0683539476	\N	76
85	2025-10-24 09:04:08.383058+00	2025-10-24 10:15:46.539219+00	\N	Pauline	Guilbaud 	\N	0689051654	\N	85
86	2025-10-24 21:18:33.376938+00	2025-10-24 21:18:33.376938+00	\N	christelle	HUIBAN	\N	0649588505	\N	86
87	2025-10-24 21:18:33.381714+00	2025-10-24 21:18:33.381714+00	\N	Nicolas	MONTAVONT	\N	\N	\N	87
90	2025-10-26 08:37:08.304575+00	2025-10-26 08:37:08.304575+00	\N	Viki	Villeneuve	\N	0638781842	\N	90
89	2025-10-26 07:42:41.777236+00	2025-10-26 18:45:11.282826+00	\N	Maëliss 	Monbon	\N	0783804201	\N	89
109	2026-05-05 19:37:11.315846+00	2026-05-05 19:37:11.315846+00	\N	Séverine 	DESILLE 	\N	\N	\N	109
91	2025-10-26 12:04:46.188246+00	2025-10-26 22:16:27.949577+00	\N	Victor	Blanchard	\N	\N	\N	91
72	2025-10-15 07:36:22.428941+00	2025-10-29 07:00:42.625535+00	\N	Nathalie 	Delvincourt 	\N	0699526971	\N	72
92	2025-10-29 12:55:23.568094+00	2025-10-29 12:55:23.568094+00	\N	Cyrille	Prevosteau	\N	\N	\N	92
110	2026-05-07 18:13:01.318917+00	2026-05-07 18:13:01.318917+00	\N	Yann	Le Doare 	\N	0651031144	\N	110
111	2026-05-08 15:44:37.442397+00	2026-05-08 15:44:37.442397+00	\N	Catherine	Jakubiec	\N	0663497345	\N	111
101	2025-11-09 20:36:50.944922+00	2025-11-09 20:36:50.944922+00	\N	Marie-Florine 	Dambakizi 	\N	0767935251	\N	101
112	2026-05-08 15:45:36.747955+00	2026-05-08 15:45:36.747955+00	\N	Zoé	CAREMEL	\N	\N	\N	112
88	2025-10-25 12:46:55.122482+00	2025-11-10 19:09:46.084043+00	\N	Gaela	Cochennec	\N	0610491382	\N	88
95	2025-11-07 10:03:14.541507+00	2025-11-12 09:03:46.293627+00	\N	Jean	Tanguy	\N	\N	\N	95
104	2025-11-14 16:59:34.357547+00	2025-11-14 16:59:34.357547+00	\N	Theo	Clemenceau 	\N	\N	\N	104
116	2026-05-25 14:27:31.063613+00	2026-05-25 14:27:31.06877+00	\N	Guide	Test	\N	\N	\N	116
94	2025-11-07 10:03:14.53648+00	2025-11-14 23:38:13.905537+00	\N	Clémence	Hillion	\N	0680361083	\N	94
100	2025-11-08 17:29:07.212345+00	2025-11-20 06:29:18.35038+00	\N	Victor 	Gaudard 	\N	\N	\N	100
103	2025-11-14 16:59:34.354951+00	2025-11-20 07:43:57.093644+00	\N	charlotte 	fourchon	\N	\N	\N	103
96	2025-11-07 20:15:02.172037+00	2025-11-20 11:24:51.898735+00	\N	marie	Bonnardot 	\N	\N	\N	96
99	2025-11-08 17:29:07.208551+00	2025-11-20 12:08:03.078596+00	\N	Émilie 	Gillier 	\N	\N	\N	99
97	2025-11-08 14:32:37.794132+00	2025-11-26 05:48:04.409557+00	\N	Virginie	Vincent	\N	0688728475	\N	97
93	2025-10-31 09:45:53.923071+00	2025-11-26 09:06:50.200159+00	\N	Emilie	Massard	\N	\N	\N	93
98	2025-11-08 14:32:37.7989+00	2025-11-26 15:37:33.478598+00	\N	Mélanie	lepretre	\N	\N	\N	98
102	2025-11-09 20:56:04.108944+00	2025-11-28 19:24:22.93164+00	\N	Sophie	Lemarié 	\N	0664215771	\N	102
107	2026-04-27 10:33:04.573893+00	2026-04-27 10:33:16.329021+00	\N	Brenda	CRESTEL	\N	\N	\N	107
108	2026-04-28 21:49:01.926452+00	2026-04-30 16:16:18.570765+00	\N	Dorine 	Fourgaut 	\N	0659738640	\N	108
106	2026-04-14 18:07:27.487247+00	2026-05-01 08:51:34.925084+00	\N	Alain 	viel 	\N	\N	\N	106
117	2026-05-31 13:52:28.404276+00	2026-06-03 06:56:46.345014+00	\N	Claire 	lebouc 	\N	0607273277	\N	117
131	2026-06-27 03:28:00.405024+00	2026-06-27 03:31:57.275467+00	\N	Yassine	Amar	\N	0673468216	\N	131
114	2026-05-18 16:02:53.550467+00	2026-06-03 07:05:23.205516+00	\N	Agathe	Séné	\N	\N	\N	114
113	2026-05-18 16:00:04.846361+00	2026-06-03 07:08:50.693732+00	\N	Anaëlle	Morel	\N	0638863382	\N	113
1	2025-08-13 10:01:49.984483+00	2026-06-29 08:42:01.563573+00	\N	Vera	LORENZETI	\N	\N	\N	1
118	2026-06-04 18:14:05.374823+00	2026-06-04 18:14:34.766187+00	\N	Eddy	Boite	\N	0784020211	\N	118
127	2026-06-17 12:00:30.833018+00	2026-07-01 07:24:32.731553+00	\N	Léa	RION	\N	\N	\N	127
121	2026-06-11 18:42:03.950616+00	2026-06-11 23:29:55.666457+00	\N	Virginie 	Chérel	\N	0670126368	\N	121
123	2026-06-12 13:54:45.809459+00	2026-06-12 13:55:30.994918+00	\N	Guide	Test	\N	\N	\N	123
115	2026-05-25 14:27:12.940348+00	2026-07-02 11:01:17.455352+00	\N	Guide	Test	\N	\N	\N	115
128	2026-06-17 12:01:57.352149+00	2026-07-02 12:39:17.012526+00	\N	Thomas	VOISIN	\N	0669439869	\N	128
105	2026-04-14 18:04:25.213463+00	2026-07-02 14:00:37.243522+00	\N	anne	viel	\N	\N	\N	105
122	2026-06-11 20:54:03.377631+00	2026-07-02 19:58:10.883793+00	\N	Jacques 	tixier	\N	0663492358	\N	122
124	2026-06-15 13:37:53.345067+00	2026-06-15 13:37:53.345067+00	\N	Maéva 	Cadeau 	\N	0619980199	\N	124
126	2026-06-17 08:12:31.661885+00	2026-06-17 08:12:31.661885+00	\N	Benoît 	Rolland 	\N	\N	\N	126
119	2026-06-05 10:24:45.515674+00	2026-07-03 11:03:50.746382+00	\N	Emilie	Bourgeois	\N	0647167599	\N	119
125	2026-06-17 08:11:50.247756+00	2026-06-17 08:26:19.729789+00	\N	ana	Rolland 	\N	\N	\N	125
129	2026-06-17 20:20:32.082912+00	2026-06-17 20:20:32.082912+00	\N	CELINE 	GUERIN 	\N	0684159138	\N	129
45	2025-10-03 19:50:51.507838+00	2026-06-18 07:58:25.736562+00	\N	Maéva	CADEAU	\N	0619980199	\N	45
130	2026-06-18 20:02:31.559529+00	2026-06-18 20:02:31.559529+00	\N	Jade	Guérin 	\N	0684159138	\N	130
120	2026-06-07 20:49:26.599504+00	2026-06-24 14:08:07.636416+00	\N	Damien 	Closier 	\N	0631788884	\N	120
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.user_roles (user_id, role_id) FROM stdin;
1	1
12	2
74	2
107	2
115	2
116	2
115	3
118	2
123	2
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: balade
--

COPY public.users (id, created_at, updated_at, deleted_at, email, authentication_code, authentication_expire_at) FROM stdin;
54	2025-10-06 16:04:32.486188+00	2025-10-28 05:44:46.530909+00	\N	adline.leon@gmail.com	905659	2025-10-28 05:49:12.324318+00
39	2025-10-03 12:08:10.603386+00	2025-10-08 05:52:34.205189+00	\N	julieglemot5@gmail.com	533442	2025-10-08 05:57:09.442132+00
2	2025-09-11 15:11:47.345709+00	2025-09-11 15:11:47.345709+00	\N	vlnb12311@gmail.com	\N	\N
4	2025-09-25 17:32:54.817709+00	2025-09-25 17:32:54.817709+00	\N	isabellecheval8@gmail.com	\N	\N
17	2025-10-01 09:46:43.606656+00	2025-10-08 06:53:50.745552+00	\N	clotilde.philippe@univ-rennes.fr	128293	2025-10-08 06:58:27.283539+00
55	2025-10-07 06:57:29.462265+00	2025-10-17 06:11:25.265265+00	\N	nadegecorbe3576@gmail.com	887999	2025-10-17 06:14:38.363563+00
6	2025-09-28 18:15:07.371909+00	2025-09-28 18:15:07.371909+00	\N	vandepeutte.coline@hotmail.fr	\N	\N
10	2025-09-30 11:18:17.150891+00	2025-09-30 11:18:17.150891+00	\N	julientrubat86@gmail.com	\N	\N
11	2025-09-30 11:18:17.155241+00	2025-09-30 11:18:17.155241+00	\N	oltraperrine@gmail.com	\N	\N
47	2025-10-04 19:25:30.473233+00	2025-10-15 06:36:19.183822+00	\N	mickael.dardaillon@gmail.com	512265	2025-10-15 06:40:57.53879+00
12	2025-09-30 21:06:25.737839+00	2025-09-30 21:06:46.903742+00	\N	valentinvllegendre@gmail.com	940737	2025-09-30 21:11:25.739041+00
13	2025-09-30 21:45:49.625917+00	2025-09-30 21:45:49.625917+00	\N	s.porcelli92@gmail.com	\N	\N
14	2025-10-01 04:21:56.775543+00	2025-10-01 04:21:56.775543+00	\N	marchand.anne@wanadoo.fr	\N	\N
18	2025-10-01 09:46:43.609608+00	2025-10-01 09:46:43.609608+00	\N	sacha.pa@free.fr	\N	\N
19	2025-10-01 09:47:42.878641+00	2025-10-01 09:47:42.878641+00	\N	francoise.razanamaro@univ-rennes.fr	\N	\N
20	2025-10-01 09:47:42.881444+00	2025-10-01 09:47:42.881444+00	\N	barbedettedamien@yahoo.fr	\N	\N
21	2025-10-01 10:28:52.250647+00	2025-10-01 10:28:52.250647+00	\N	stephylou35@hotmail.fr	\N	\N
22	2025-10-01 10:31:10.588195+00	2025-10-01 10:31:42.823339+00	\N	laulo561@gmail.com	974070	2025-10-01 10:36:42.822937+00
37	2025-10-03 12:01:54.384802+00	2025-10-08 08:11:21.533717+00	\N	edith.blin@inria.fr	703746	2025-10-08 08:15:21.100795+00
16	2025-10-01 06:56:19.032403+00	2025-10-01 11:06:55.776981+00	\N	jpg.guillot@gmail.com	655046	2025-10-01 11:11:22.661994+00
23	2025-10-01 16:25:00.23256+00	2025-10-01 16:25:00.23256+00	\N	shauezarx@mozmail.com	\N	\N
24	2025-10-01 16:25:00.235344+00	2025-10-01 16:25:00.235344+00	\N	thomas.anberree@gmail.com	\N	\N
26	2025-10-01 19:03:30.205334+00	2025-10-01 19:03:30.205334+00	\N	delphinebechetoille@gmail.com	\N	\N
27	2025-10-01 19:04:54.605643+00	2025-10-01 19:04:54.605643+00	\N	kersual35@gmail.com	\N	\N
34	2025-10-03 08:20:04.850123+00	2025-10-03 08:20:04.850123+00	\N	bezier.marion@outlook.fr	\N	\N
35	2025-10-03 08:20:04.853667+00	2025-10-03 08:20:04.853667+00	\N	walker-35@hotmail.fr	\N	\N
36	2025-10-03 11:13:10.857724+00	2025-10-03 11:13:10.857724+00	\N	amarie.guillot@laposte.net	\N	\N
38	2025-10-03 12:08:10.600359+00	2025-10-03 12:08:10.600359+00	\N	guyot.rebecca@gmail.com	\N	\N
42	2025-10-03 16:46:20.553607+00	2025-10-03 16:46:20.553607+00	\N	stephanie.lelabousse@gmail.com	\N	\N
43	2025-10-03 16:47:24.20326+00	2025-10-03 16:47:24.20326+00	\N	ch.berthault@gmail.com	\N	\N
61	2025-10-10 09:01:47.030288+00	2025-10-10 09:01:47.030288+00	\N	soizic.masson@orange.fr	\N	\N
7	2025-09-29 06:17:58.55879+00	2025-10-05 06:21:44.227168+00	\N	soizzzm@yahoo.fr	894678	2025-10-05 06:26:19.902312+00
48	2025-10-05 10:59:30.059873+00	2025-10-05 10:59:30.059873+00	\N	erwan.vieville@gmail.com	\N	\N
49	2025-10-05 10:59:30.062769+00	2025-10-05 10:59:30.062769+00	\N	rosalba.vieville@gmail.com	\N	\N
50	2025-10-05 10:59:30.064261+00	2025-10-05 10:59:30.064261+00	\N	yvon.vieville@gmail.com	\N	\N
30	2025-10-02 08:19:41.269825+00	2025-10-19 07:22:37.817571+00	\N	ann.dauphin@laposte.net	850073	2025-10-19 07:27:14.212327+00
52	2025-10-05 14:42:12.967942+00	2025-10-05 14:42:12.967942+00	\N	canellethomas@gmail.com	\N	\N
53	2025-10-05 14:42:12.969504+00	2025-10-05 14:42:12.969504+00	\N	rondel.charlotte@gmail.com	\N	\N
46	2025-10-04 19:25:30.469457+00	2025-10-15 06:41:07.942006+00	\N	gaelletwk1@yahoo.fr	243037	2025-10-15 06:45:46.213416+00
41	2025-10-03 15:53:36.489344+00	2025-11-09 15:32:49.750276+00	\N	pyfab.balcon@wanadoo.fr	627067	2025-11-09 15:37:20.914111+00
28	2025-10-01 19:28:56.248568+00	2025-10-15 07:22:55.239784+00	\N	jeanyvesnath@gmail.com	648479	2025-10-15 07:27:26.496276+00
31	2025-10-02 12:49:19.729182+00	2025-10-19 13:55:56.230665+00	\N	isabelle.lins@univ-rennes.fr	055771	2025-10-19 14:00:56.230482+00
32	2025-10-02 13:49:50.157218+00	2025-10-07 08:12:36.102735+00	\N	catheline.barbara@gmail.com	199219	2025-10-07 08:17:06.315928+00
44	2025-10-03 19:31:11.68597+00	2025-10-07 08:36:09.199287+00	\N	pauline.trion@gmail.com	084554	2025-10-07 08:38:40.804458+00
58	2025-10-07 10:18:06.793426+00	2025-10-07 10:18:06.793426+00	\N	lennysaxo@gmail.com	\N	\N
74	2025-10-16 09:06:39.68039+00	2025-10-16 09:07:16.882384+00	\N	edith.merdrignac@univ-rennes1.fr	288403	2025-10-16 09:11:39.683734+00
70	2025-10-13 22:02:19.165191+00	2025-10-15 11:50:22.411841+00	\N	apollineprivat@gmail.com	075783	2025-10-15 11:55:12.092294+00
62	2025-10-11 17:40:12.390403+00	2026-05-25 14:24:09.98979+00	\N	victordenis01@gmail.com	151024	2026-05-25 14:29:09.989397+00
84	2025-10-23 11:36:24.950645+00	2025-10-23 11:36:24.950645+00	\N	cauchy.lucie@hotmail.fr	\N	\N
15	2025-10-01 05:25:04.073833+00	2025-10-21 07:43:04.71002+00	\N	yannick.robert35@gmail.com	231483	2025-10-21 07:47:44.211636+00
83	2025-10-22 09:48:58.616258+00	2025-10-22 09:48:58.616258+00	\N	beatice.rabault@gmail.com	\N	\N
29	2025-10-01 19:28:56.251491+00	2025-10-15 12:26:48.047517+00	\N	muller.nath8@gmail.com	621032	2025-10-15 12:31:21.137878+00
77	2025-10-19 17:52:18.682102+00	2025-10-19 17:57:31.383871+00	\N	demeulenaere.manon@yahoo.fr	183634	2025-10-19 18:02:04.439904+00
51	2025-10-05 14:27:26.389436+00	2025-10-17 05:06:34.25432+00	\N	r.delannay@orange.fr	606956	2025-10-17 05:11:00.567526+00
66	2025-10-13 17:22:50.709835+00	2025-10-19 17:42:33.070959+00	\N	laura.giommi@live.fr	051253	2025-10-19 17:47:08.483009+00
78	2025-10-20 08:11:18.574313+00	2025-10-20 08:11:18.574313+00	\N	malika.teneur@gmail.com	\N	\N
60	2025-10-09 18:33:00.146334+00	2025-10-19 06:15:14.924964+00	\N	cyrilpinchon@gmail.com	871543	2025-10-19 06:19:45.084685+00
67	2025-10-13 17:24:48.719858+00	2025-10-19 17:44:33.818392+00	\N	solene.barbe@hotmail.fr	761350	2025-10-19 17:48:47.442486+00
57	2025-10-07 08:15:51.328361+00	2025-10-20 21:04:32.630692+00	\N	pascal.navarre@wanadoo.fr	224566	2025-10-20 21:09:00.837096+00
64	2025-10-12 10:05:10.134011+00	2025-10-20 08:50:01.566093+00	\N	nfeurgard@gmail.com	405163	2025-10-20 08:54:36.207044+00
79	2025-10-20 10:03:27.815839+00	2025-10-21 07:46:15.591886+00	\N	d_lecocq@hotmail.com	836710	2025-10-21 07:50:56.128135+00
80	2025-10-21 11:32:11.43315+00	2025-10-21 11:32:11.43315+00	\N	estelle.clua@gmail.com	\N	\N
56	2025-10-07 08:12:46.86652+00	2025-10-20 21:05:57.619814+00	\N	mart.piel@orange.fr	927572	2025-10-20 21:10:28.599219+00
40	2025-10-03 15:13:00.79973+00	2025-10-22 09:50:26.779429+00	\N	beatrice.rabault@gmail.com	711194	2025-10-22 09:54:55.186804+00
81	2025-10-21 11:57:00.251259+00	2025-10-21 11:57:00.251259+00	\N	diane.schmitt3@wanadoo.fr	\N	\N
59	2025-10-07 14:06:22.259192+00	2025-10-21 15:32:34.172266+00	\N	iven.lelouedec@gmail.com	468325	2025-10-21 15:37:17.526328+00
65	2025-10-12 10:05:10.1398+00	2025-10-22 06:41:39.021344+00	\N	isabelle.brendlen@icloud.com	137721	2025-10-22 06:46:10.240079+00
68	2025-10-13 18:20:09.172515+00	2025-10-22 11:34:06.411626+00	\N	daheronf@gmail.com	432531	2025-10-22 11:38:36.919357+00
8	2025-09-29 14:46:53.062136+00	2025-10-24 05:03:35.461085+00	\N	lolia.cozette@gmail.com	251511	2025-10-24 05:08:10.095699+00
76	2025-10-18 09:48:07.788937+00	2025-10-24 04:58:58.96966+00	\N	anthonyambroise68@gmail.com	433166	2025-10-24 05:03:58.969372+00
73	2025-10-15 16:49:09.775807+00	2025-10-24 16:44:38.188002+00	\N	davoctau@gmail.com	467923	2025-10-24 16:49:09.635326+00
69	2025-10-13 22:00:05.960517+00	2025-10-24 09:56:06.588015+00	\N	gael-yann@hotmail.fr	976828	2025-10-24 10:00:23.530267+00
85	2025-10-24 09:04:08.381447+00	2025-10-24 10:15:46.538343+00	\N	pauline-guilbaud@hotmail.fr	369239	2025-10-24 10:20:28.489415+00
33	2025-10-02 18:27:29.733058+00	2025-10-24 16:48:28.983576+00	\N	elodie.octau@gmail.com	975100	2025-10-24 16:53:02.265039+00
86	2025-10-24 21:18:33.37618+00	2025-10-24 21:18:33.37618+00	\N	christellehuiban@gmail.com	\N	\N
87	2025-10-24 21:18:33.381318+00	2025-10-24 21:18:33.381318+00	\N	nicolas@montavont.net	\N	\N
90	2025-10-26 08:37:08.303756+00	2025-10-26 08:37:08.303756+00	\N	viki.villeneuve@hotmail.fr	\N	\N
25	2025-10-01 19:03:30.202584+00	2025-10-26 09:07:31.096721+00	\N	jeanne.lichou@gmail.com	724265	2025-10-26 09:12:01.797134+00
82	2025-10-21 12:09:25.374801+00	2025-10-26 11:43:30.712482+00	\N	titouan.millon3@gmail.com	001669	2025-10-26 11:47:56.238394+00
5	2025-09-28 18:15:07.368053+00	2025-10-28 07:08:18.709434+00	\N	Vandepeutte.coline@hotmail.fr	946774	2025-10-28 07:13:03.490858+00
63	2025-10-12 09:22:56.359098+00	2025-10-29 12:53:51.584859+00	\N	romainprevosteau@gmail.com	556810	2025-10-29 12:58:34.714595+00
72	2025-10-15 07:36:22.428754+00	2025-10-29 07:00:42.624876+00	\N	nathalie.strugalski@hotmail.fr	977447	2025-10-29 07:05:14.273611+00
9	2025-09-29 21:11:56.235923+00	2025-11-08 06:26:32.331918+00	\N	mariepierreremeur@gmail.com	416924	2025-11-08 06:31:02.94858+00
88	2025-10-25 12:46:55.121745+00	2025-11-10 19:09:46.083507+00	\N	gaela_c@protonmail.com	050578	2025-11-10 19:14:12.125541+00
3	2025-09-25 16:46:09.746199+00	2025-11-14 07:38:24.796129+00	\N	ecologique.pro@gmail.com	056844	2025-11-14 07:43:24.795696+00
89	2025-10-26 07:42:41.776455+00	2025-10-26 18:45:11.282226+00	\N	maeliss.monbon@lilo.org	608618	2025-10-26 18:49:43.172821+00
91	2025-10-26 12:04:46.187347+00	2025-10-26 22:16:27.948811+00	\N	viblanchar@gmail.com	710409	2025-10-26 22:21:15.690713+00
92	2025-10-29 12:55:23.567034+00	2025-10-29 12:55:23.567034+00	\N	cyrille.prevosteau@sfr.fr	\N	\N
75	2025-10-17 16:43:42.969218+00	2025-10-29 19:09:14.083372+00	\N	alexia.muzas@gmail.com	281758	2025-10-29 19:13:56.589463+00
45	2025-10-03 19:50:51.507269+00	2026-06-18 07:58:25.735463+00	\N	maeva.cadeau@orange.fr	933047	2026-06-18 07:58:57.686828+00
71	2025-10-14 19:48:53.312985+00	2025-10-29 19:10:02.148703+00	\N	axa_mzs_work@outlook.fr	764645	2025-10-29 19:14:46.836093+00
101	2025-11-09 20:36:50.944692+00	2025-11-09 20:36:50.944692+00	\N	mf.dambakizi@gmail.com	\N	\N
95	2025-11-07 10:03:14.541081+00	2025-11-12 09:03:46.292634+00	\N	j.tanguy.22@gmail.com	712993	2025-11-12 09:07:01.047355+00
104	2025-11-14 16:59:34.357357+00	2025-11-14 16:59:34.357357+00	\N	theo.clemenceau352@gmail.com	\N	\N
94	2025-11-07 10:03:14.535722+00	2025-11-14 23:38:13.905298+00	\N	hillion.c@gmail.com	371039	2025-11-14 23:42:54.739672+00
100	2025-11-08 17:29:07.211935+00	2025-11-20 06:29:18.349522+00	\N	victor.15156@gmail.com	288669	2025-11-20 06:34:01.826018+00
103	2025-11-14 16:59:34.354551+00	2025-11-20 07:43:57.093459+00	\N	charlotte.fourchon@gmail.com	284587	2025-11-20 07:48:39.42773+00
96	2025-11-07 20:15:02.171602+00	2025-11-20 11:24:51.897894+00	\N	Marie.bonnardot@outlook.com	493165	2025-11-20 11:29:09.258745+00
99	2025-11-08 17:29:07.20804+00	2025-11-20 12:08:03.077657+00	\N	emilie.gillier@lilo.org	745251	2025-11-20 12:12:40.960656+00
97	2025-11-08 14:32:37.793202+00	2025-11-26 05:48:04.40933+00	\N	vincent.virginie@gmail.com	657444	2025-11-26 05:52:37.528818+00
93	2025-10-31 09:45:53.922742+00	2025-11-26 09:06:50.199268+00	\N	emiliemassard@wanadoo.fr	322336	2025-11-26 09:11:01.613554+00
98	2025-11-08 14:32:37.798433+00	2025-11-26 15:37:33.477843+00	\N	contact@papi-jean.com	979920	2025-11-26 15:42:13.669171+00
102	2025-11-09 20:56:04.108789+00	2025-11-28 19:24:22.930816+00	\N	y.lemarie@sfr.fr	627020	2025-11-28 19:29:00.685516+00
107	2026-04-27 10:33:04.573013+00	2026-04-27 10:33:16.328392+00	\N	brenda.crestel@univ-rennes.fr	793468	2026-04-27 10:38:04.578196+00
108	2026-04-28 21:49:01.925055+00	2026-04-30 16:16:18.570125+00	\N	fourgautdorine@yahoo.fr	039123	2026-04-30 16:20:55.719852+00
106	2026-04-14 18:07:27.486308+00	2026-05-01 08:51:34.924483+00	\N	vieldelangon@wanadoo.fr	959672	2026-05-01 08:54:55.22841+00
109	2026-05-05 19:37:11.314443+00	2026-05-05 19:37:11.314443+00	\N	redcorvette2018@gmail.com	\N	\N
110	2026-05-07 18:13:01.318505+00	2026-05-07 18:13:01.318505+00	\N	yann@linuxconsole.org	\N	\N
111	2026-05-08 15:44:37.441457+00	2026-05-08 15:44:37.441457+00	\N	catherinejaku@hotmail.com	\N	\N
112	2026-05-08 15:45:36.747212+00	2026-05-08 15:45:36.747212+00	\N	zoe.caremel@mailo.com	\N	\N
116	2026-05-25 14:27:31.062566+00	2026-05-25 14:27:31.067902+00	\N	contact@ecologique.pro	143563	2026-05-25 14:32:31.067334+00
117	2026-05-31 13:52:28.39987+00	2026-06-03 06:56:46.342243+00	\N	claire.freezone@gmail.com	268916	2026-06-03 07:01:46.341714+00
128	2026-06-17 12:01:57.351075+00	2026-07-02 12:39:17.011416+00	\N	baladeeco.retold540@passmail.net	332024	2026-07-02 12:43:46.086261+00
114	2026-05-18 16:02:53.549541+00	2026-06-03 07:05:23.204342+00	\N	agathesene0@gmail.com	711781	2026-06-03 07:10:05.432215+00
113	2026-05-18 16:00:04.844411+00	2026-06-03 07:08:50.692394+00	\N	agathe.senedu56@gmail.com	788568	2026-06-03 07:13:28.345184+00
105	2026-04-14 18:04:25.210907+00	2026-07-02 14:00:37.242288+00	\N	vielderennes@orange.fr	880051	2026-07-02 14:05:03.182976+00
118	2026-06-04 18:14:05.370783+00	2026-06-04 18:14:34.765151+00	\N	eddy.boite@me.com	995028	2026-06-04 18:19:05.389939+00
122	2026-06-11 20:54:03.376443+00	2026-07-02 19:58:10.882656+00	\N	jtixier@free.fr	008371	2026-07-02 20:02:46.492426+00
119	2026-06-05 10:24:45.51317+00	2026-07-03 11:03:50.745262+00	\N	emibouu@gmail.com	670184	2026-07-03 11:08:36.969525+00
121	2026-06-11 18:42:03.94931+00	2026-06-11 23:29:55.665292+00	\N	amelaint@yahoo.fr	251700	2026-06-11 23:34:21.610517+00
123	2026-06-12 13:54:45.806975+00	2026-06-12 13:55:30.993613+00	\N	valfrotinapro@proton.me	457803	2026-06-12 14:00:30.993162+00
124	2026-06-15 13:37:53.342703+00	2026-06-15 13:37:53.342703+00	\N	maeva.cadeau@orange.fr 	\N	\N
126	2026-06-17 08:12:31.66077+00	2026-06-17 08:12:31.66077+00	\N	breizhbrj@yahoo.fr	\N	\N
125	2026-06-17 08:11:50.245302+00	2026-06-17 08:26:19.728701+00	\N	rollandj.anais@gmail.com	910577	2026-06-17 08:30:53.805295+00
129	2026-06-17 20:20:32.080446+00	2026-06-17 20:20:32.080446+00	\N	guerice1@gmail.com	\N	\N
130	2026-06-18 20:02:31.558244+00	2026-06-18 20:02:31.558244+00	\N	Jade35.guerin@gmail.com	\N	\N
120	2026-06-07 20:49:26.598287+00	2026-06-24 14:08:07.635336+00	\N	Closier.damien@postel.bzh	661430	2026-06-24 14:12:47.065046+00
131	2026-06-27 03:28:00.402267+00	2026-06-27 03:31:57.274354+00	\N	2.yassine.amar@gmail.com	472911	2026-06-27 03:36:38.667604+00
1	2025-08-13 10:01:49.984483+00	2026-06-29 08:42:01.562316+00	\N	contact@baladeecologique.com	830917	2026-06-29 08:45:53.193674+00
127	2026-06-17 12:00:30.831902+00	2026-07-01 07:24:32.730381+00	\N	lea.rion310@gmail.com	575899	2026-07-01 07:29:17.790366+00
115	2026-05-25 14:27:12.937866+00	2026-07-02 11:01:17.454132+00	\N	valfortinapro@proton.me	613664	2026-07-02 11:05:43.052967+00
\.


--
-- Name: addresses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.addresses_id_seq', 1, false);


--
-- Name: guides_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.guides_id_seq', 6, true);


--
-- Name: payments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.payments_id_seq', 67, true);


--
-- Name: permissions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.permissions_id_seq', 1, false);


--
-- Name: ramble_guides_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.ramble_guides_id_seq', 275, true);


--
-- Name: ramble_prices_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.ramble_prices_id_seq', 394, true);


--
-- Name: ramble_registration_groups_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.ramble_registration_groups_id_seq', 37, true);


--
-- Name: ramble_registrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.ramble_registrations_id_seq', 197, true);


--
-- Name: rambles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.rambles_id_seq', 42, true);


--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.roles_id_seq', 1, false);


--
-- Name: seeds_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.seeds_id_seq', 8, true);


--
-- Name: user_permission_overrides_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.user_permission_overrides_id_seq', 1, false);


--
-- Name: user_profiles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.user_profiles_id_seq', 131, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: balade
--

SELECT pg_catalog.setval('public.users_id_seq', 131, true);


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);


--
-- Name: guides guides_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.guides
    ADD CONSTRAINT guides_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: ramble_guides ramble_guides_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_guides
    ADD CONSTRAINT ramble_guides_pkey PRIMARY KEY (id);


--
-- Name: ramble_prices ramble_prices_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_prices
    ADD CONSTRAINT ramble_prices_pkey PRIMARY KEY (id);


--
-- Name: ramble_registration_groups ramble_registration_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registration_groups
    ADD CONSTRAINT ramble_registration_groups_pkey PRIMARY KEY (id);


--
-- Name: ramble_registrations ramble_registrations_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registrations
    ADD CONSTRAINT ramble_registrations_pkey PRIMARY KEY (id);


--
-- Name: rambles rambles_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.rambles
    ADD CONSTRAINT rambles_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: seeds seeds_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.seeds
    ADD CONSTRAINT seeds_pkey PRIMARY KEY (id);


--
-- Name: guides uni_guides_email; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.guides
    ADD CONSTRAINT uni_guides_email UNIQUE (email);


--
-- Name: payments uni_payments_stripe_payment_intent_id; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT uni_payments_stripe_payment_intent_id UNIQUE (stripe_payment_intent_id);


--
-- Name: permissions uni_permissions_name; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT uni_permissions_name UNIQUE (name);


--
-- Name: roles uni_roles_name; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT uni_roles_name UNIQUE (name);


--
-- Name: seeds uni_seeds_name; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.seeds
    ADD CONSTRAINT uni_seeds_name UNIQUE (name);


--
-- Name: user_profiles uni_user_profiles_user_id; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT uni_user_profiles_user_id UNIQUE (user_id);


--
-- Name: users uni_users_email; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_email UNIQUE (email);


--
-- Name: user_permission_overrides user_permission_overrides_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_permission_overrides
    ADD CONSTRAINT user_permission_overrides_pkey PRIMARY KEY (id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_addresses_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_addresses_deleted_at ON public.addresses USING btree (deleted_at);


--
-- Name: idx_guides_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_guides_deleted_at ON public.guides USING btree (deleted_at);


--
-- Name: idx_guides_user_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE UNIQUE INDEX idx_guides_user_id ON public.guides USING btree (user_id);


--
-- Name: idx_payments_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_payments_deleted_at ON public.payments USING btree (deleted_at);


--
-- Name: idx_payments_group_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_payments_group_id ON public.payments USING btree (group_id);


--
-- Name: idx_payments_guide_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_payments_guide_id ON public.payments USING btree (guide_id);


--
-- Name: idx_payments_registration_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_payments_registration_id ON public.payments USING btree (registration_id);


--
-- Name: idx_payments_stripe_charge_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_payments_stripe_charge_id ON public.payments USING btree (stripe_charge_id);


--
-- Name: idx_payments_stripe_payment_intent_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_payments_stripe_payment_intent_id ON public.payments USING btree (stripe_payment_intent_id);


--
-- Name: idx_permissions_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_permissions_deleted_at ON public.permissions USING btree (deleted_at);


--
-- Name: idx_ramble_guides_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_guides_deleted_at ON public.ramble_guides USING btree (deleted_at);


--
-- Name: idx_ramble_guides_guide_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_guides_guide_id ON public.ramble_guides USING btree (guide_id);


--
-- Name: idx_ramble_guides_ramble_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_guides_ramble_id ON public.ramble_guides USING btree (ramble_id);


--
-- Name: idx_ramble_prices_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_prices_deleted_at ON public.ramble_prices USING btree (deleted_at);


--
-- Name: idx_ramble_registration_groups_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_registration_groups_deleted_at ON public.ramble_registration_groups USING btree (deleted_at);


--
-- Name: idx_ramble_registration_groups_ramble_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_registration_groups_ramble_id ON public.ramble_registration_groups USING btree (ramble_id);


--
-- Name: idx_ramble_registrations_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_registrations_deleted_at ON public.ramble_registrations USING btree (deleted_at);


--
-- Name: idx_ramble_registrations_group_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_registrations_group_id ON public.ramble_registrations USING btree (group_id);


--
-- Name: idx_ramble_registrations_ramble_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_registrations_ramble_id ON public.ramble_registrations USING btree (ramble_id);


--
-- Name: idx_ramble_registrations_user_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_ramble_registrations_user_id ON public.ramble_registrations USING btree (user_id);


--
-- Name: idx_rambles_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_rambles_deleted_at ON public.rambles USING btree (deleted_at);


--
-- Name: idx_rambles_payment_guide_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_rambles_payment_guide_id ON public.rambles USING btree (payment_guide_id);


--
-- Name: idx_roles_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_roles_deleted_at ON public.roles USING btree (deleted_at);


--
-- Name: idx_seeds_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_seeds_deleted_at ON public.seeds USING btree (deleted_at);


--
-- Name: idx_user_permission_overrides_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_user_permission_overrides_deleted_at ON public.user_permission_overrides USING btree (deleted_at);


--
-- Name: idx_user_permission_overrides_permission_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_user_permission_overrides_permission_id ON public.user_permission_overrides USING btree (permission_id);


--
-- Name: idx_user_permission_overrides_user_id; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_user_permission_overrides_user_id ON public.user_permission_overrides USING btree (user_id);


--
-- Name: idx_user_profiles_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_user_profiles_deleted_at ON public.user_profiles USING btree (deleted_at);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: balade
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: payments fk_payments_group; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT fk_payments_group FOREIGN KEY (group_id) REFERENCES public.ramble_registration_groups(id);


--
-- Name: payments fk_payments_guide; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT fk_payments_guide FOREIGN KEY (guide_id) REFERENCES public.guides(id);


--
-- Name: payments fk_payments_registration; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT fk_payments_registration FOREIGN KEY (registration_id) REFERENCES public.ramble_registrations(id);


--
-- Name: ramble_guides fk_ramble_guides_guide; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_guides
    ADD CONSTRAINT fk_ramble_guides_guide FOREIGN KEY (guide_id) REFERENCES public.guides(id);


--
-- Name: ramble_guides fk_ramble_guides_ramble; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_guides
    ADD CONSTRAINT fk_ramble_guides_ramble FOREIGN KEY (ramble_id) REFERENCES public.rambles(id);


--
-- Name: ramble_registration_groups fk_ramble_registration_groups_ramble; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registration_groups
    ADD CONSTRAINT fk_ramble_registration_groups_ramble FOREIGN KEY (ramble_id) REFERENCES public.rambles(id);


--
-- Name: ramble_registrations fk_ramble_registration_groups_registrations; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registrations
    ADD CONSTRAINT fk_ramble_registration_groups_registrations FOREIGN KEY (group_id) REFERENCES public.ramble_registration_groups(id);


--
-- Name: ramble_registrations fk_ramble_registrations_ramble; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registrations
    ADD CONSTRAINT fk_ramble_registrations_ramble FOREIGN KEY (ramble_id) REFERENCES public.rambles(id);


--
-- Name: ramble_registrations fk_ramble_registrations_user; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registrations
    ADD CONSTRAINT fk_ramble_registrations_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: rambles fk_rambles_payment_guide; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.rambles
    ADD CONSTRAINT fk_rambles_payment_guide FOREIGN KEY (payment_guide_id) REFERENCES public.guides(id);


--
-- Name: ramble_prices fk_rambles_prices; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_prices
    ADD CONSTRAINT fk_rambles_prices FOREIGN KEY (ramble_id) REFERENCES public.rambles(id);


--
-- Name: ramble_registrations fk_rambles_registrations; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.ramble_registrations
    ADD CONSTRAINT fk_rambles_registrations FOREIGN KEY (ramble_id) REFERENCES public.rambles(id);


--
-- Name: role_permissions fk_role_permissions_permission; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES public.permissions(id);


--
-- Name: role_permissions fk_role_permissions_role; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: user_permission_overrides fk_user_permission_overrides_permission; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_permission_overrides
    ADD CONSTRAINT fk_user_permission_overrides_permission FOREIGN KEY (permission_id) REFERENCES public.permissions(id);


--
-- Name: user_profiles fk_user_profiles_address; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT fk_user_profiles_address FOREIGN KEY (address_id) REFERENCES public.addresses(id);


--
-- Name: user_roles fk_user_roles_role; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: user_roles fk_user_roles_user; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: guides fk_users_guide; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.guides
    ADD CONSTRAINT fk_users_guide FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_permission_overrides fk_users_permission_overrides; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_permission_overrides
    ADD CONSTRAINT fk_users_permission_overrides FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_profiles fk_users_user_profile; Type: FK CONSTRAINT; Schema: public; Owner: balade
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT fk_users_user_profile FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

