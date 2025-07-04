import {getCobalt} from "./cobalt.ts";
import type {Metadata} from "./api.ts";

async function runCmd(cmd: string[]): Promise<number> {
    const ffmpeg = Bun.spawn({
        cmd: cmd,
        stdout: 'inherit',
        stderr: 'inherit',
    });

    return await ffmpeg.exited;
}

export async function dl(id: string, metadata: Metadata): Promise<void> {
    const data = await getCobalt(id);

    console.log(data);

    const fixedFile = `./dl/${id}/audio.m4a`;

    const audioCmd = [
        'ffmpeg',
        '-y',
        '-i', data.url,
        '-i', metadata.thumbnail, '-map', '0', '-map', '1', '-disposition:v:0', 'attached_pic', // set cover
        '-c:a', 'aac',
        '-vn',
        '-t', metadata.length.toString(),
        '-metadata', `title="${metadata.title}"`,
        '-metadata', `artist="${metadata.authors}"`,
        fixedFile,
    ];

    console.log(audioCmd.join(' '));

    const audio = await runCmd(audioCmd);

    console.log(`${id} audio downloaded`);

    const hlsCmd = [
        'ffmpeg',
        '-y',
        '-i', fixedFile,
        '-c:a', 'aac',
        '-f', 'hls',
        '-vn',
        '-t', metadata.length.toString(),
        '-hls_playlist_type', 'vod',
        '-hls_segment_filename', `./dl/${id}/segment_%03d.ts`,
        `./dl/${id}/hls.m3u8`
    ];

    console.log(hlsCmd.join(' '));

    const hls = await runCmd(hlsCmd);

    console.log(`${id}: audio - ${audio}, hls - ${hls}`);
}