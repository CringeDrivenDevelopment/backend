import {Innertube, UniversalCache} from "youtubei.js";
// @ts-ignore
import {MusicInlineBadge} from "youtubei.js/dist/src/parser/nodes";

const innertube = await Innertube.create({
    lang: 'en',
    location: 'US',
    retrieve_player: false,
    enable_safety_mode: false,
    generate_session_locally: false,
    enable_session_cache: true,
    device_category: 'desktop',
    cookie: '',
    cache: new UniversalCache(
        true,
        './cache'
    )
});

export interface Track extends Metadata {
    id: string;
    explicit: boolean;
}

export interface Metadata {
    length: number;
    title: string;
    authors: string;
    thumbnail: string;
}

export async function search(query: string): Promise<Awaited<Track>[]> {
    const search = await innertube.music.search(decodeURI(query), {
        type: 'song'
    });

    if (!search.songs) {
        return [];
    }

    return await Promise.all(search.songs.contents.map(async song => {
        const authors = song.artists?.map(x => x.name).join(', ')!;
        const thumbnail = song.thumbnail?.contents[song.thumbnails!.length - 1]?.url!;
        const info = await innertube.getInfo(song.id!);
        const length = info.basic_info.duration!;
        const explicit = song.badges?.find(item => {
            const badge = item as MusicInlineBadge;
            return badge.icon_type === 'MUSIC_EXPLICIT_BADGE';
        }) !== undefined;

        return {
            id: song.id!,
            title: song.title!,
            authors: authors,
            thumbnail: thumbnail,
            length: length,
            explicit: explicit,
        };
    }));
}

export async function getMetadata(id: string): Promise<Metadata> {
    const song = await innertube.music.getInfo(id);

    const authors = song.basic_info.author ?? '';
    const thumbnail = song.basic_info.thumbnail?.at(song.basic_info.thumbnail!.length - 1)!.url!;
    const info = await innertube.getInfo(id);
    const length = info.basic_info.duration!;

    return {
        title: song.basic_info.title!,
        authors: authors,
        thumbnail: thumbnail,
        length: length,
    };
}