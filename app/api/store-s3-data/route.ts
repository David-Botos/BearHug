import { createClient } from '@supabase/supabase-js';
import { Database } from '@/database.types';
import { constructS3Directory } from '@/utils/supabase/constructS3Directory';

export async function POST(request: Request) {
  console.log('🔵 /api/store-s3-data API route called');
  
  try {
    const { roomUrl } = await request.json();
    console.log('📥 Received roomUrl:', roomUrl);
    
    const url = process.env.NEXT_PUBLIC_SUPABASE_URL;
    const serviceKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;
    
    if (!url || !serviceKey) {
      console.error('❌ Missing Supabase credentials');
      return new Response('Missing Supabase credentials', { status: 500 });
    }
    
    console.log('🔑 Initializing Supabase client with URL:', url);
    const supabase = createClient<Database>(url, serviceKey);
    
    console.log('📂 Constructing S3 directory from roomUrl');
    const s3_directory = constructS3Directory(roomUrl);
    console.log('📁 Generated S3 directory:', s3_directory);

    console.log('💾 Attempting to insert data into Supabase');
    const { error } = await supabase.from('calls').insert({
      room_url: roomUrl,
      s3_folder_dir: s3_directory,
      created_at: new Date().toISOString(),
    });

    if (error) {
      console.error('❌ Supabase insert error:', error);
      throw error;
    }

    console.log('✅ Successfully stored data in Supabase');
    return new Response('Success', { status: 200 });
  } catch (error) {
    console.error('🔴 Error in API route:', error);
    return new Response(`Failed to store data: ${error instanceof Error ? error.message : 'Unknown error'}`, { status: 500 });
  }
}