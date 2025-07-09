import {getCobalt} from "./cobalt.ts";
import type {Metadata} from "./api.ts";
import { mkdir } from "fs/promises";
import JSZip from "jszip";

async function runCmd(cmd: string[]): Promise<number> {
    const ffmpeg = Bun.spawn({
        cmd: cmd,
        stdout: 'ignore',
        stderr: 'inherit',
    });

    return await ffmpeg.exited;
}

export async function archive(songs: string[], archiveName: string): Promise<void> {
    const jszip = new JSZip();
    const folder = jszip.folder('songs');

    await Promise.all(songs.map(async (song) => {
        const audioFile = Bun.file(`./dl/${song}/audio.m4a`);
        const fileName = await Bun.file(`./dl/${song}/FILENAME`).text();
        const arrayBuffer = await audioFile.arrayBuffer();

        folder!.file(fileName, arrayBuffer);
    }));

    const content = await jszip.generateAsync({
        type: 'nodebuffer',
        compression: 'DEFLATE',
        compressionOptions: { level: 9 }
    });
    await Bun.write(`./dl/${archiveName}`, content);
}

export async function dl(cobaltUrl: string, id: string, metadata: Metadata): Promise<void> {
    const data = await getCobalt(cobaltUrl, id);

    await mkdir(`./dl/${id}`, { recursive: true });

    await Bun.write(`./dl/${id}/FILENAME`, `[${id}] ${metadata.authors} - ${metadata.title}.m4a`)

    const fixedFile = `./dl/${id}/audio.m4a`;

    const audioCmd = [
        'ffmpeg',
        '-y',
        '-i', data.url,
        '-i', metadata.thumbnail, '-map', '0:a', '-map', '1:v', '-disposition:v:0', 'attached_pic', // set cover
        '-c:a', 'aac',
        '-c:v', 'copy',
        '-t', metadata.length.toString(),
        '-metadata', `title=${metadata.title.replaceAll(' ', '\ ')}`,
        '-metadata', `artist=${metadata.authors.replaceAll(' ', '\ ')}`,
        fixedFile,
    ];

    console.log(audioCmd.join(' '));

    const audio = runCmd(audioCmd);

    const hlsCmd = [
        'ffmpeg',
        '-y',
        '-i', data.url,
        '-c:a', 'aac',
        '-f', 'hls',
        '-vn',
        '-ss', (Math.floor(metadata.length / 2)).toString(),
        '-t', '10',
        '-hls_playlist_type', 'vod',
        '-hls_segment_filename', `./dl/${id}/segment_%03d.ts`,
        `./dl/${id}/hls.m3u8`
    ];

    console.log(hlsCmd.join(' '));

    const hls = runCmd(hlsCmd);

    await Promise.all([audio, hls]);
}